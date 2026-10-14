package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
)

type createAdminArgs struct{ email, name string }

func parseCreateAdmin(args []string) (createAdminArgs, error) {
	fs := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	email := fs.String("email", "", "e-mail do admin")
	name := fs.String("name", "", "nome de exibicao (padrao: parte do e-mail antes do @)")
	if err := fs.Parse(args); err != nil {
		return createAdminArgs{}, fmt.Errorf("argumentos invalidos: %w (use: create-admin --email voce@exemplo.com [--name Nome])", err)
	}
	if *email == "" {
		return createAdminArgs{}, errors.New("informe --email (use: create-admin --email voce@exemplo.com [--name Nome])")
	}
	return createAdminArgs{email: *email, name: *name}, nil
}

func createAdmin(ctx context.Context, args []string, out io.Writer) error {
	in, err := parseCreateAdmin(args)
	if err != nil {
		return err
	}
	pool, err := openDatabase(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()
	uc := accounts.CreateAdmin{Store: postgres.NewAccounts(pool), PublicURL: os.Getenv("APP_PUBLIC_URL")}
	invite, err := uc.Run(ctx, in.email, in.name)
	if err != nil {
		return err
	}
	verb := "criado"
	if invite.Reissued {
		verb = "ja existia; convite reemitido"
	}
	_, err = fmt.Fprintf(out, "Admin %s: %s\nConvite (uso unico, vale ate %s): %s\n",
		verb, invite.Email, invite.ExpiresAt.Local().Format(time.TimeOnly), invite.URL)
	return err
}
