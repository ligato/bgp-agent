## Synchronizer Utility Package

Contains synchronizer utility which waits until counter matches 
an expected number during an specific time.

```golang
counter := uint32(0)
expected := uint32(1)
duration := 1 * time.Second

err := WaitCounterMatch(duration, expected, &counter)
```
