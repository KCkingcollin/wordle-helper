# A simple(ish) wordle helper

## It currently has 2 modes

### mode 1: simple cli arguments

In simple cli mode you provide it with anywhere from 1 to 3 arguments when calling the program.

Argument 1: Type phrase, use _ in the place of any unknown characters before the search term  
Argument 2: In the second argument type any characters you know are in the word, but not the position  
Use _ to put the characters in the place you know they are not in  
Argument 3: Type in any characters you know aren't in the word  
Make sure the characters are all lowercase  

You can get that same syntax by just typing --help or -h in the first argument.

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

