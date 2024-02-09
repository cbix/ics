package main

import (
	"html/template"
	"io"
)

var indexTpl = template.Must(template.New("index").Parse(`
<!doctype html>
<html lang="en">
	<head>
		<title>ICS Adapters</title>
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.min.css">
	</head>
	<body>
		<h1>ICS Adapters</h1>
		<p>Please visit the <a href="https://github.com/cbix/ics">GitHub repo</a> for more information, source code and support!</p>
		<ul>
			{{ range . }}
			<li>
				<a href="{{ .Url }}" target="_blank">{{ .Name }}</a>
				(<a href="{{ .Name }}.ics">ics</a>):
				{{ .Description }}
			</li>
			{{ end }}
		</ul>
	</body>
</html>
`))

func writeIndex(w io.Writer) error {
	return indexTpl.Execute(w, adapters)
}
