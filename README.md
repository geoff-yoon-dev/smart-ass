# smart-ass

This project was inspired by the [the fuck](https://github.com/nvbn/thefuck) project.
GPT was combined with the [the fuck] project to make the rule base, the fuck, more flexible and able to accommodate a variety of commands. Incorrect commands written by the user will be corrected through GPT.

![alt text](https://github.com/geoff-yoon-dev/smart-ass/blob/main/docs/images/smartass_demo.png?raw=true)


## requirement

### setting bash_history update
```
echo -e "\
function share_history\n\
  history -a\n\
  history -c\n\
  history -r\n\
}\n\
shopt -u histappend\n\
PROMPT_COMMAND=\"share_history; $PROMPT_COMMAND\"" >> ~/.bashrc
```

### setting openai api key
```
export OPENAI_KEY="your openai key"
```

## install
```
wget https://github.com/geoff-yoon-dev/smart-ass/releases/download/v0.0.1/smartass
```

## build 
```
go build -0 smartass
```

### path setting
```
export PATH=$PATH:/your/smartass/path
```

## run
```
smartass
```
if you want to exec
```
smartass -x
```