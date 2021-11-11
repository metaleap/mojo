
userName := .env("USER")


main :=
	\ i := 0 < 10
		myProg userName
		i = i+1
	<- 0


myProg := (user)
	? .strEq(user, "_")
		user = "king of castle"
	| .strEq("root")
		user = "wannabe"
	| .strEq("root")
		.never

	.print (.env("GOPATH").len == 0) ? "Hello, " | "Hello, Gopher ",
	.print user, "!\n"
