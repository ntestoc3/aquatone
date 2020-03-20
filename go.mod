module aquatone

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/lair-framework/go-nmap v0.0.0-20191202052157-3507e0b03523 // indirect
	github.com/michenriksen/aquatone/agents v1.7.0 
	github.com/michenriksen/aquatone/core v1.7.0 
	github.com/michenriksen/aquatone/parsers v1.7.0 
	github.com/mvdan/xurls v1.1.0 // indirect
	github.com/parnurzeal/gorequest v0.2.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/michenriksen/aquatone/core => ./core

replace github.com/michenriksen/aquatone/parsers => ./parsers

replace github.com/michenriksen/aquatone/agents => ./agents

