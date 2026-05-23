package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/astrolink/node/internal/config"
)

var setupGroupOrder = []string{"system", "security", "database", "payments", "opennds"}

func main() {
	defaultPath := config.FromEnv().AstrolinkEnvFile
	envPath := flag.String("env-file", defaultPath, "caminho do arquivo .env local")
	flag.Parse()

	if err := runSetup(os.Stdin, os.Stdout, strings.TrimSpace(*envPath)); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}

func runSetup(input io.Reader, output io.Writer, envPath string) error {
	if envPath == "" {
		envPath = ".env"
	}
	file, err := config.LoadEnvFile(envPath)
	if err != nil {
		return fmt.Errorf("ler %s: %w", envPath, err)
	}

	status := config.BuildSetupStatus(file)
	reader := bufio.NewReader(input)
	values := map[string]string{}

	fmt.Fprintln(output, "Astrolink setup local")
	fmt.Fprintf(output, "Arquivo: %s\n", envPath)
	fmt.Fprintln(output, "Pressione Enter para manter valores atuais. Segredos configurados nao sao exibidos.")

	for _, groupKey := range setupGroupOrder {
		group, ok := status.Groups[groupKey]
		if !ok {
			continue
		}
		fmt.Fprintf(output, "\n[%s]\n", group.Label)
		for _, field := range group.Fields {
			answer, err := askField(reader, output, field)
			if err != nil {
				return err
			}
			if answer != "" {
				values[field.Key] = answer
			}
		}
	}

	if len(values) == 0 {
		fmt.Fprintln(output, "\nNenhuma alteracao aplicada.")
		return nil
	}
	if err := config.ApplySetupPatch(file, values); err != nil {
		return err
	}
	if err := config.SaveEnvFileAtomic(envPath, file); err != nil {
		return fmt.Errorf("gravar %s: %w", envPath, err)
	}
	fmt.Fprintln(output, "\nConfiguracao salva. Reinicie o node para carregar as novas variaveis.")
	return nil
}

func askField(reader *bufio.Reader, output io.Writer, field config.SetupField) (string, error) {
	current := field.Value
	if field.Secret && field.Configured {
		current = "configurado"
	}
	if current != "" {
		fmt.Fprintf(output, "%s [%s]: ", field.Label, current)
	} else {
		fmt.Fprintf(output, "%s: ", field.Label)
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
