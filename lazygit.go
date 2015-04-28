package main

import (
    "os"
    "strings"
    "os/exec"
    "fmt"
    "github.com/codegangsta/cli"
)

func RunCommand(cmd string) (string, error) {
    c := exec.Command("sh","-c", cmd)
    if output, err := c.CombinedOutput(); err != nil {
        return "",err
    } else {
        return string(output),nil
    }
}

func SetGitConfig(key string, value string) error {
    cmd := fmt.Sprintf("git config %v '%v'", key, value)

    if _,err := RunCommand(cmd); err != nil {
        return err
    } else {
        return nil
    }
}

func GetGitConfig(key string) (string,error) {
    cmd := fmt.Sprintf("git config %v", key)

    if output,err := RunCommand(cmd); err != nil {
        return "",err
    } else {
        return output,nil
    }
}


func OnMasterBranch() (bool,error) {
    if output,err := RunCommand("git status"); err != nil {
        return false,err
    } else {

        if master,err := GetGitConfig("lazygit.master"); err != nil {
            return false,err
        } else {
            return strings.Contains(output,"On branch " + master),nil
        }
    }
}

func GetIgnoreBranches() ([]string,error) {

    if ignoreBranches,err := GetGitConfig("lazygit.ignorebranches"); err != nil {
        return []string{},err
    } else {
        return strings.Split(ignoreBranches,","),err
    }

}


func GetMergedBranches() ([]string,error) {
    if output,err := RunCommand("git branch --merged"); err != nil {
        return []string{},err
    } else {
        mergedBranches := make([]string, 0)
        ignoreBranches,_ := GetIgnoreBranches()

        outputLines := strings.Split(output,"\n")
        for _,element := range outputLines {

            element = strings.TrimSpace(element)

            if strings.HasPrefix(element,"*") {
                element = strings.TrimSpace(element[1:])
            }

            found := false
            for _,ignoreBranch := range ignoreBranches {
                if ignoreBranch == element {
                    found = true
                    break
                }
            }

            if !found {
                mergedBranches = append(mergedBranches,element)
            }

        }

        return mergedBranches,nil
    }
}

func main() {
    app := cli.NewApp()
    app.Name = "lazygit"
    app.Usage = "Git helpers for lazy gits"
    app.Version = "0.1.0"
    app.Author = "Will Ogden"

    app.Commands = []cli.Command{
        {
            Name: "cleanup",
            Aliases: []string{"c"},
            Usage: "removes merged local branches and prunes remotes",
            Action: func(c *cli.Context) {

                if value,err := OnMasterBranch(); err != nil || !value {
                    println("oops, you're not on the master branch!")
                    return
                }

                if mergedBranches,err:= GetMergedBranches(); err != nil {
                    println("oops, something went wrong", err.Error())
                } else {

                    if len(mergedBranches) > 0 {
                        mergedBranchesJoined := strings.Join(mergedBranches, " ")

                        cmd := fmt.Sprintf("git branch -d %v", mergedBranchesJoined)

                        if output,err := RunCommand(cmd); err != nil {
                            println("oops, branches couldn't be deleted! ",err.Error())
                        } else {
                            println(output)
                        }

                    }

                }

                if output,err := RunCommand("git remote prune origin"); err != nil {
                    println("oops, remote prune failed! ",err.Error())
                } else {
                    println(output)
                }

            },
        },
        {
            Name: "ignorebranches",
            Aliases: []string{"i"},
            Usage: "don't cleanup ignored branches",
            Action: func(c *cli.Context) {
                ignoreBranches := strings.Join(c.Args(),",")
                if err := SetGitConfig("lazygit.ignorebranches",ignoreBranches); err != nil {
                    println("oops, something went wrong", err.Error())
                } else {
                    println("branches ignored: ", ignoreBranches)
                }
            },
        },
        {
            Name: "setmaster",
            Aliases: []string{"m"},
            Usage: "set the master branch",
            Action: func(c *cli.Context) {

                if err := SetGitConfig("lazygit.master",c.Args().First()); err != nil {
                    println("oops, something went wrong", err.Error())
                }

            },
        },
    }

    app.Run(os.Args)
}
