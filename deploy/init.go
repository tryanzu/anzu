package main

import (
        "bytes"
        "golang.org/x/crypto/ssh"
        "fmt"
        "io/ioutil"
)

func main() {

        key, err := getKeyFile()

        if err != nil {
                panic(err)
        }

        config := &ssh.ClientConfig{
            User: "root",
            Auth: []ssh.AuthMethod{
                ssh.PublicKeys(key),
            },
        }

        hosts := []string{"sp-service1.spartangeek.com:22", "sp-service2.spartangeek.com:22",}

        for _, host := range hosts {

                client, err := ssh.Dial("tcp", host, config)

                if err != nil {
                    panic("Failed to dial: " + err.Error())
                }

                session, err := client.NewSession()

                if err != nil {
                        panic("Failed to create session: " + err.Error())
                }

                defer session.Close()

                var b bytes.Buffer
                session.Stdout = &b
                if err := session.Run("cd /opt/go/src/github.com/fernandez14/spartangeek-board && git pull && service sp-board restart"); err != nil {
                        panic("Failed to run: " + err.Error())
                }

                fmt.Println(b.String())
        }
}

func getKeyFile() (key ssh.Signer, err error){
    file := "/root/.ssh/id_rsa"
    buf, err := ioutil.ReadFile(file)
    if err != nil {
        return
    }
    key, err = ssh.ParsePrivateKey(buf)
    if err != nil {
        return
     }
    return
}
