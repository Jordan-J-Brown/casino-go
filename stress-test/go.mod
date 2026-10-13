module github.com/Jordan-J-Brown/casino-go/stress-test

go 1.27.1

replace github.com/Jordan-J-Brown/casino-go/casino => ../casino

replace github.com/Jordan-J-Brown/casino-go/engines/mines => ../engines/mines

replace github.com/Jordan-J-Brown/casino-go/pkg => ../pkg

require (
	github.com/Jordan-J-Brown/casino-go/engines/mines v0.0.0-20261005175900-651c4fdd3bd8
	github.com/Jordan-J-Brown/casino-go/pkg v0.0.0-20261013150336-7a9fc33a63f2
	github.com/alecthomas/kong v1.16.1
	github.com/shopspring/decimal v1.5.0
)
