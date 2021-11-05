# Fire

Fetch an object from Firebase RTDB.

## Usage

To use `fire`, create two environment variables with the following data:

1. Set `FIRE_ACCOUNT` to the fully qualified path to a serviceAccount.json file
2. Set `FIRE_URL` to the database url `https://your-app.firebaseio.com`

To fetch an object:

```bash
    fire [-p] [-s] path/to/obj`
    
    Flags
      -k shallow, one key per line
      -p pretty print
      -s fetch shallow
```

### NOTES

* The sort order from `-k` is indeterminate (a feature of Go maps), but you could always pipe the results to sort: `fire -k accounts | sort`.

## Build

If adding features or just playing around:

```bash
go run main.go -p path/to/some/object
```

Install to your GOPATH

```bash
go build -o ~/go/bin/fire main.go
```

## Examples

Once installed, just run fire:

```bash
fire -p account/types

fire -s versions

fire -k some/deep/nested/object 
```