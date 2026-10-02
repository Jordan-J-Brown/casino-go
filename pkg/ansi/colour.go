package ansi

const (
	ESC                  = "\033" //Escape Sequence
	CSI_SGR              = "["    //Control Sequence Introducer - Select Graphic Rendition
	SGR_End              = "m"
	Reset                = ESC + CSI_SGR + "0" + SGR_End
	Bold                 = ESC + CSI_SGR + "1" + SGR_End
	Faint                = ESC + CSI_SGR + "2" + SGR_End
	Italic               = ESC + CSI_SGR + "3" + SGR_End
	Underline            = ESC + CSI_SGR + "4" + SGR_End
	SetForegroundBlack   = ESC + CSI_SGR + "30" + SGR_End
	SetForegroundRed     = ESC + CSI_SGR + "31" + SGR_End
	SetForegroundGreen   = ESC + CSI_SGR + "32" + SGR_End
	SetForegroundYellow  = ESC + CSI_SGR + "33" + SGR_End
	SetForegroundBlue    = ESC + CSI_SGR + "34" + SGR_End
	SetForegroundMagenta = ESC + CSI_SGR + "35" + SGR_End
	SetForegroundCyan    = ESC + CSI_SGR + "36" + SGR_End
	SetForegroundWhite   = ESC + CSI_SGR + "37" + SGR_End
)
