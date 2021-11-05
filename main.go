package main

import (
	"bytes"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
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
		DatabaseURL: env("FIRE_URL"), //"https://your-app.firebaseio.com",
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

	} else {
		var data interface{}

		if shallow {
			if err := ref.GetShallow(ctx, &data); err != nil {
				return "", errors.Wrap(err, "Error reading from database")
			}
		} else {
			if err := ref.Get(ctx, &data); err != nil {
				return "", errors.Wrap(err, "Error reading from database")
			}
		}

		buf, err := json.Marshal(data)
		if err != nil {
			return "", errors.Wrap(err, "Error marshalling data")
		}

		if prettyPrint {
			return jsonPrettyPrint(buf), nil
		}
		return string(buf), nil
	}
}

func jsonPrettyPrint(buf []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, buf, "", "   ")
	if err != nil {
		return string(buf)
	}
	return out.String()
}

func env(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		fmt.Fprintf(os.Stderr, "Environment variable %s must be set\n", key)
	}
	return val
}
