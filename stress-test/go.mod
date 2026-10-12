module github.com/Jordan-J-Brown/casino-go/stress-test

go 1.27.1

replace github.com/Jordan-J-Brown/casino-go/casino => ../casino

replace github.com/Jordan-J-Brown/casino-go/engines/mines => ../engines/mines

replace github.com/Jordan-J-Brown/casino-go/pkg => ../pkg

require (
	github.com/Jordan-J-Brown/casino-go/casino v0.0.0-20261005175900-651c4fdd3bd8
	github.com/Jordan-J-Brown/casino-go/engines/mines v0.0.0-20261005175900-651c4fdd3bd8
	github.com/Jordan-J-Brown/casino-go/pkg v0.0.0-20261005175900-651c4fdd3bd8
	github.com/shopspring/decimal v1.5.0
)

require github.com/alecthomas/kong v1.16.1 // indirect
