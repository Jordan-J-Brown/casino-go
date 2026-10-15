package logsx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/ansi"
	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
)

type TextHandler struct {
	*slog.TextHandler
	workingDir string
	w          io.Writer
	buffPool   sync.Pool
}

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) *TextHandler {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wd, err = filepath.EvalSymlinks(wd)
	if err != nil {
		panic(err)
	}

	handler := &TextHandler{
		w:           w,
		TextHandler: slog.NewTextHandler(w, opts),
		workingDir:  wd,
	}
	handler.buffPool.New = func() any {
		return bytes.NewBuffer(nil)
	}

	return handler
}

func (t *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	// TODO: add "{package_name}/{file}:{column}" to log.

	//{timestamp} {coloured_level} {correlation_id} {msg} {args...}

	buff := t.buffPool.Get().(*bytes.Buffer)
	defer func() {
		buff.Reset()
		t.buffPool.Put(buff)
	}()

	//Time
	buff.WriteString(r.Time.Format(time.DateTime))
	buff.WriteByte(' ')

	//Level
	writeLevel(buff, r.Level)
	buff.WriteByte(' ')

	cid := contextx.CorrelationIdFromContext(ctx)
	if cid != "" {
		buff.WriteString("cid=")
		buff.WriteString(cid)
		buff.WriteByte(' ')
	}

	buff.WriteString(r.Message)
	buff.WriteByte(' ')

	buff.WriteString("args={")

	i := 0
	numAttrs := r.NumAttrs()
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(buff, "%s=%v", a.Key, a.Value)
		if i < numAttrs-1 {
			buff.WriteByte(' ')
		}
		i++
		return true
	})

	buff.WriteByte('}')
	buff.WriteByte('\n')

	_, err := t.w.Write(buff.Bytes())

	return err
}

func writeLevel(b *bytes.Buffer, level slog.Level) {
	b.WriteString(ansi.Bold)
	switch level {
	case slog.LevelDebug:
		b.WriteString(ansi.SetForegroundBlue)
	case slog.LevelInfo:
		b.WriteString(ansi.SetForegroundGreen)
	case slog.LevelWarn:
		b.WriteString(ansi.SetForegroundYellow)
	case slog.LevelError:
		b.WriteString(ansi.SetForegroundRed)
	}
	b.WriteString(level.String())
	b.WriteString(ansi.Reset)
}
