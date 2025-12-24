package amigo

const sqlTemplate = `{{.UpAnnotation}}{{if .Transactional}} tx=true{{else}} tx=false{{end}}


{{.DownAnnotation}}{{if .Transactional}} tx=true{{else}} tx=false{{end}}

`
