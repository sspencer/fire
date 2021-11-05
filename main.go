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
	"strings"
)

func main() {
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

	json, err := fetch(flag.Arg(0), *keyPtr, *prettyPtr, *shallowPtr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(json)
}

func fetch(obj string, keyPrint, prettyPrint, shallow bool) (string, error) {
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

	ref := client.NewRef(obj)

	if keyPrint {
		return keyFetch(ctx, ref)
	} else if shallow {
		return shallowFetch(ctx, ref, prettyPrint)
	} else {
		return deepFetch(ctx, ref, prettyPrint)
	}
}

func keyFetch(ctx context.Context, ref *db.Ref) (string, error) {
	var data map[string]bool
	if err := ref.GetShallow(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	keys := make([]string, len(data))
	i := 0
	for k := range data {
		keys[i] = k
		i++
	}

	return strings.Join(keys, "\n"), nil
}

func shallowFetch(ctx context.Context, ref *db.Ref, prettyPrint bool) (string, error) {
	var data interface{}
	if err := ref.GetShallow(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	return jsonPrettyPrint(data, prettyPrint)
}

func deepFetch(ctx context.Context, ref *db.Ref, prettyPrint bool) (string, error) {
	var data interface{}
	if err := ref.Get(ctx, &data); err != nil {
		return "", errors.Wrap(err, "Error reading from database")
	}

	return jsonPrettyPrint(data, prettyPrint)
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
