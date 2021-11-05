package main

import (
	"bytes"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"os"
	"path"
	"strings"
)

type cmdParams struct {
	obj string
	append bool
	key bool
	pretty bool
	shallow bool
}

func main() {
	appendPtr := flag.Bool("a", false, "Prepend path to keys [-k]")
	keyPtr := flag.Bool("k", false, "Print top level keys, one per line")
	prettyPtr := flag.Bool("p", false, "Pretty print JSON")
	shallowPtr := flag.Bool("s", false, "Shallow Fetch")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Fetch firebase object")
		fmt.Printf("Usage: %s [-p] object\n", os.Args[0])
		os.Exit(1)
	}

	cmd := cmdParams{
		obj:     flag.Arg(0),
		append:  *appendPtr,
		key:     *keyPtr,
		pretty:  *prettyPtr,
		shallow: *shallowPtr,
	}

	json, err := fetch(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(json)
}

func fetch(cmd cmdParams) (string, error) {
	ctx := context.Background()
	conf := &firebase.Config{
		DatabaseURL: env("FIRE_URL"),
	}

	opt := option.WithCredentialsFile(env("FIRE_ACCOUNT"))

	// Initialize the app with a service account, granting admin privileges
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return "", errors.Wrap(err, "Error initializing app")
	}

	client, err := app.Database(ctx)
	if err != nil {
		return "", errors.Wrap(err, "Error initializing database client")
	}

	ref := client.NewRef(cmd.obj)

	if cmd.key {
		return keyFetch(ctx, ref, cmd)
	} else if cmd.shallow {
		return shallowFetch(ctx, ref, cmd)
	} else {
		return deepFetch(ctx, ref, cmd)
	}
}

func keyFetch(ctx context.Context, ref *db.Ref, cmd cmdParams) (string, error) {
	var data map[string]bool
	if err := ref.GetShallow(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	keys := make([]string, len(data))
	i := 0
	for k := range data {
		var key string
		if cmd.append {
			key = path.Join(cmd.obj, k)
		} else {
			key = k
		}
		keys[i] = key
		i++
	}

	return strings.Join(keys, "\n"), nil
}

func shallowFetch(ctx context.Context, ref *db.Ref, cmd cmdParams) (string, error) {
	var data interface{}
	if err := ref.GetShallow(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	return jsonPrettyPrint(data, cmd.pretty)
}

func deepFetch(ctx context.Context, ref *db.Ref, cmd cmdParams) (string, error) {
	var data interface{}
	if err := ref.Get(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	return jsonPrettyPrint(data, cmd.pretty)
}

func jsonPrettyPrint(data interface{}, prettyPrint bool) (string, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "Error marshalling data")
	}

	if prettyPrint {
		var out bytes.Buffer
		err = json.Indent(&out, buf, "", "   ")
		if err != nil {
			return "", errors.Wrap(err, "Error formatting JSON")
		}
		return out.String(), nil
	}

	return string(buf), nil
}

func env(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		fmt.Fprintf(os.Stderr, "Environment variable %s must be set\n", key)
		os.Exit(1)
	}
	return val
}
