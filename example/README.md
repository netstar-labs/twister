# twister examples

| Example | What it shows | Run |
|---|---|---|
| [permute](permute/main.go) | generating the full permutation set for a brand, grouped by fuzzer, plus `tld-swap` with a caller-supplied TLD list | `go run ./example/permute` |

Build standalone with `GOWORK=off` if the surrounding workspace doesn't list this
module.

For the CLI over labels from arguments or stdin, see
[../docs/userguide.md](../docs/userguide.md):

```sh
go run ./app/twister permute paypal                    # every technique, one variant per line
go run ./app/twister permute -f homoglyph,omission paypal
printf 'paypal\ngoogle\n' | go run ./app/twister permute -tld com,net
```

The [../diff/](../diff/) module cross-validates twister against `twist` — every
edit-based variant is detected as a distance-1 near-miss.
