package template

import (
	"bytes"
	"context"
	"github.com/keikoproj/alert-manager/pkg/log"
	"html/template"
)

//ProcessTemplate process the go lang template byb substituting with the provided values
func ProcessTemplate(ctx context.Context, input string, val map[string]string) (string, error) {
	log := log.Logger(ctx, "internal.template", "ProcessTemplate")
	log.V(4).Info("processing template", "input", input)

	tmpl, err := template.New("alert.tmpl").Parse(input)
	if err != nil {
		log.Error(err, "template is NOT valid")
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, val); err != nil {
		log.Error(err, "unable to execute template")
		return "", err
	}
	log.Info("Template executed successfully", "temp", buf.String())
	return buf.String(), nil
}
