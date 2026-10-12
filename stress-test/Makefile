games:=
buildTags:=
run-mines:
	@go run -tags="${buildTags}" ./cmd/cli/ run mines ${games} \
		--workers=10 \
		--logs.level=debug \
		--game.type=mines \
		--game.args=mines=3 \
		--game.args=positions=0,1,2,3 \
		--game.args=seed=yT/xmMMkzsstxDca0Evf968ZXbeOZKaKZZ0nqiXpfVU=

