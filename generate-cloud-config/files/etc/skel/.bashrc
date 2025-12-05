# Custom bash configuration
export EDITOR=vim
export VISUAL=vim

# Custom aliases
alias ll='ls -lah'
alias update='sudo apt-get update && sudo apt-get upgrade'

# Custom prompt
PS1='\[\\033[01;32m\]\\u@\h\[\\033[00m\]:\[\\033[01;34m\]\w\[\\033[00m\]\$ '