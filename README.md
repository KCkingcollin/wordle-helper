# A simple(ish) wordle helper

## It currently has 2 modes

### mode 1: simple cli arguments

In simple cli mode you provide it with anywhere from 1 to 3 arguments when calling the program.

Argument 1: Known character locations, use _ in the place of any unknown characters before the last known character    
Argument 2: Characters known to be in the word, but not the position, use _ in the place of possible positions  
Argument 3: Characters you know aren't in the word  
Sending a number as an argument sets the size of the word, otherwise it will default to 5  

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

### mode 2: raw cli (work in progress)

If you run it without any arguments then it'll run in cli mode by default (still working on it, and currently does not work).  

The advantage of cli mode is that its much faster (though argument mode is pretty fast anyway), and I have plans to automatically keep the last input the user typed in the input to make it easier to use.

I realize this mode will be largely unused, however I plan to use a similar interface in other projects, so I wanted to go ahead and play with it a bit.

## Sorting

In both modes the output is sorted based on the popularity of the word, this makes picking the right word a lot easier.

## Building

Go is extremely easy to use, all you have to do is insure go is installed (and git if you clone the repo), then simply run `go build wordle-helper.go` inside the downloaded repo.  
Here's how you'd do it in linux as an example (It should also work in windows the same way, but keep in mind I did not make this with windows in mind):  
```
git clone https://github.com/KCkingcollin/wordle-helper
cd wordle-helper
go build wordle-helper.go
```

Once I finish up the raw cli mode I'll most likely create a release with a linux binary, and a windows exe if it doesn't give me any trouble.

