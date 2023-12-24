package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type FixCommandResponse struct {
	IsIncorrect      bool   `json:"is_incorrect,omitempty"`
	IncorrectCommand string `json:"incorrect_command,omitempty"`
	IncorrectPoint   string `json:"incorrect_point,omitempty"`
	IncorrectReason  string `json:"incorrect_reason,omitempty"`
	FixedCommand     string `json:"fixed_command,omitempty"`
}

func GetShell() string {
	shellPath := os.Getenv("SHELL")
	_, shell := path.Split(shellPath)

	return shell
}

func GetRecentCommand() (string, error) {
	shell := GetShell()
	home := os.Getenv("HOME")
	shellHistory := fmt.Sprintf("%s/.%s_history", home, shell)
	cmd := exec.Command("tail", "-n 1", shellHistory)
	var outbyte, errbyte bytes.Buffer
	cmd.Stdout = &outbyte
	cmd.Stderr = &errbyte
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outbyte.String(), nil
}

func FixCommand(targetCommand string) (*FixCommandResponse, error) {
	apiKey := os.Getenv("OPENAI_KEY")

	if len(apiKey) == 0 {
		log.Fatal("[OPENAI_KEY] env variable is empty. \n Please write {export OPENAI_KEY=\"your openai key\"} on ~/.profile \n and command source ~/.profile")
	}

	systemPrompt := "" +
		"[CONTEXT]\n" +
		"Act a Linux expert.\n" +
		"If an error occurs because [USER] executes an incorrect Linux command, it is your role to correct it with the correct command.\nTo modify a command, follow the [STEP] below.\n\n" +
		"[STEP]\n" +
		"0.<evaluation command>\nDistinguishes whether this command is a correct command or an incorrect command. If the command is correct, {{is_incorrect}} is false, [STEP] 1. 2. 3. will not be performed. and\n{{incorrect_point}} says \"none\" and {{incorrect_reason}} says \"This is a correct command.\" Enter the string called\n{{fixed_command}} The entered command is still inserted.\nIf the command is incorrect, {{is_incorrect}} is true, perform [STEP] 1. 2. 3.\n 1. <incorrect_point>Find the part where the command is incorrect and surround it with **.(Example: git commmit -m \"initial commit\" => git *commmit* -m \"initial commit\")\n2. <incorrect_reason>Take your time and explain why the command is incorrect.\n3. <fixed_command>Please correct the incorrect command and change it to the correct command.\n" +
		"\n\n"

	userPrompt := "" +
		"[INCOLLECT_COMMAND]\n" +
		"%s\n" +
		"\n" +
		"[INSTUCTION]\n" +
		"Please change [INCOLLECT_COMMAND] above to the correct command.\n" +
		"\n" +
		"The output format follows the json format below.\n" +
		"\n" +
		"{" +
		"    \"is_incorrect\" : {{is_incorrect}}\n" +
		"    \"incorrect_command\" : {{incorrect_command}}\n" +
		"    \"incorrect_point\" : {{incorrect_point}}\n" +
		"    \"incorrect_reason\" : {{incorrect_reason}}\n" +
		"    \"fixed_command\" : {{fixed_command}}\n" +
		"}"

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(userPrompt, targetCommand),
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	fixCommandResponse := &FixCommandResponse{}

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &fixCommandResponse)
	if err != nil {
		return nil, err
	}

	return fixCommandResponse, nil
}

func ExecFixedCommand(fixedCommand string) (string, error) {
	command := strings.Split(fixedCommand, " ")
	cmd := exec.Command(command[0], command[1:]...)
	var outbyte, errbyte bytes.Buffer
	cmd.Stdout = &outbyte
	cmd.Stderr = &errbyte
	err := cmd.Run()

	if err != nil {
		return errbyte.String(), err
	}
	return outbyte.String(), nil
}

func main() {

	isExec := flag.Bool("x", false, "exec fixed command")
	flag.Parse()

	targetCommand, err := GetRecentCommand()
	if err != nil {
		log.Fatal(err)
	}

	fixCommandResponse, err := FixCommand(targetCommand)
	if err != nil {
		log.Fatal(err)
	}

	if !fixCommandResponse.IsIncorrect {
		fmt.Printf("command : %s is not incorrect\n", targetCommand)
	}

	fmt.Printf("command : %s", targetCommand)
	fmt.Printf("incorrect point : %s\n", fixCommandResponse.IncorrectPoint)
	fmt.Printf("incorrect reason : %s\n", fixCommandResponse.IncorrectReason)
	fmt.Printf("fixed command : %s\n", fixCommandResponse.FixedCommand)

	if *isExec {
		fmt.Printf("exec command : %s\n", fixCommandResponse.FixedCommand)
		output, _ := ExecFixedCommand(fixCommandResponse.FixedCommand)
		fmt.Println(output)
	}

}
