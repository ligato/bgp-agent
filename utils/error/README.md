### Error Utility Package

Contains a set of error related utilities 

`PanicIfError` will panic if error is present

```golang
err := bgpAgent.Init()
error.PanicIfError(err)
```