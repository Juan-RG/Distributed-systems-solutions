module t1

go 1.17

replace com => ./com

replace utils => ./utils

require com v0.0.0-00010101000000-000000000000

require (
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	utils v0.0.0-00010101000000-000000000000 // indirect
)
