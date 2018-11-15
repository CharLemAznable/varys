module github.com/CharLemAznable/varys

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20180927165925-5295e8364332
	golang.org/x/net => github.com/golang/net v0.0.0-20181102091132-c10e9556a7bc
	golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180928133829-e4b3c5e90611
	golang.org/x/text => github.com/golang/text v0.3.0
)

require (
	github.com/CharLemAznable/gcache v0.0.0-20181114094542-b3f5bfabfe11
	github.com/CharLemAznable/gql v0.0.0-20181115032521-b0c34ac5b503
	github.com/CharLemAznable/httpreq v0.0.0-20181114094433-ba039b0139de
	github.com/go-sql-driver/mysql v1.4.1
	github.com/kataras/iris v11.0.3+incompatible
)
