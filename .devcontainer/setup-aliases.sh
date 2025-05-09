#!/bin/bash
# Add aliases to shell config files

# Define your aliases
ALIASES="
alias myalias='command-to-run'
alias another='other-command'
"

# Add to bashrc if it exists
if [ -f ~/.bashrc ]; then
  echo "$ALIASES" >> ~/.bashrc
fi

# Add to zshrc if it exists
if [ -f ~/.zshrc ]; then
  echo "$ALIASES" >> ~/.zshrc
fi

# Make sure the changes take effect in current session
if [ -n "$BASH_VERSION" ]; then
  source ~/.bashrc
elif [ -n "$ZSH_VERSION" ]; then
  source ~/.zshrc
fi