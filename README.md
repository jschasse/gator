# Gator

## Program Requirements
-----------------------------------------
  - GO
  - PostgreSQL

### How to Install
-----------------------------------------
  After you have Go and PostgreSQL on your system you want to run ```go install github.com/jschasse/gator@latest```. IF you DONT have your **$GOBIN** path set, makesure to set it so that you can use gator from any command line in your system. 

#### How to add the Go bin directory to your PATH (if not already there):

Identify your Go bin directory:

1. If GOBIN is set: `echo $GOBIN`
2. If GOPATH is set: `echo $GOPATH/bin`
3. Otherwise, it's likely: `$HOME/go/bin`

Add it to your PATH: Open your shell's configuration file (e.g., ~/.bashrc, ~/.zshrc):
`nano ~/.bashrc # or ~/.zshrc, etc.`

Add the following line at the end (replace /path/to/your/go/bin with the actual path from step 1):
`export PATH=$PATH:/path/to/your/go/bin`

For example, if it's $HOME/go/bin:
`export PATH=$PATH:$HOME/go/bin`

Apply the changes: Source the configuration file or open a new terminal window:
`source ~/.bashrc # or source ~/.zshrc`

Once the directory containing gator is in your PATH, you can run it from any location by simply typing:
`gator`


