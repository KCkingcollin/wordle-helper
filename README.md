# A simple(ish) wordle helper

### Using cli arguments

Argument 1: Known character locations, use _ in the place of any unknown characters  
Argument 2: Characters known to be in the word, use _ in the place of possible positions  
Argument 3: Characters you know aren't in the word  
Sending a number as an argument sets the size of the word, otherwise it will default to 5  
_ is only required in the positions before the known characters  

You can get the argument syntax by using `--help` or `-h` in the first argument.

#### example usage:

```
wordle-helper _rate t_e_r hiowbuc
```

#### output:
```
frate
drate
prate
grate
```

## Sorting

The output is sorted based on the popularity of the word, this makes picking the right word a lot easier.

## Building

Go is extremely easy to use, all you have to do is insure go is installed (and git if you clone the repo), then simply run `go build wordle-helper.go` inside the downloaded repo.  
Here's how you'd do it in linux as an example (It should also work in windows the same way, but keep in mind I did not make this with windows in mind):  
```
git clone https://github.com/KCkingcollin/wordle-helper
cd wordle-helper
go build wordle-helper.go
```
