// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

export const CODE_EXAMPLE = {
    'Node.js': {
        langCode: 'js',
        installation: `npm i memphis-dev --save`,
        producer: `const memphis = require("memphis-dev");

(async function () {
    let memphisConn;
    try {
        memphisConn = await memphis.connect({
            host: "<memphis-host>",
            username: "<application type username>",
            connectionToken: "<connection_token>"
        });

        const producer = await memphisConn.producer({
            stationName: "<station_name>",
            producerName: "myProducer"
        });
        for (let index = 0; index < 100; index++) {
            await producer.produce({
                message: Buffer.from('Hello world')
            });
            console.log("Message sent");
        }
        console.log("All messages sent");
    } catch (ex) {
        console.log(ex);
        memphisConn && memphisConn.close();
    }
    })();`,
        consumer: `const memphis = require("memphis-dev");

(async function () {
    let memphisConn;
    try {
        memphisConn = await memphis.connect({
            host: "<memphis-host>",
            username: "<application type username>",
            connectionToken: "<connection_token>"
        });
        
        const consumer = await memphisConn.consumer({
            stationName: "<station_name>",
            consumerName: "myConsumer",
            consumerGroup: ""
        });
        consumer.on("message", message => {
            console.log(message.getData().toString());
            message.ack();
        });
        consumer.on("error", error => {
            console.log(error);
        });
    } catch (ex) {
        console.log(ex);
        memphisConn && memphisConn.close();
    }
    })();`
    },

    Typescript: {
        langCode: 'js',
        installation: `npm i memphis-dev --save`,
        producer: `import memphis from "memphis-dev"

(async function () {
    let memphisConn;
    try {
        memphisConn = await memphis.connect({
            host: "<memphis-host>",
            username: "<application type username>",
            connectionToken: "<connection_token>"
        });

        const producer = await memphisConn.producer({
            stationName: "<station_name>",
            producerName: "myProducer"
        });
        for (let index = 0; index < 100; index++) {
            await producer.produce({
                message: Buffer.from('Hello world')
            });
            console.log("Message sent");
        }
        console.log("All messages sent");
    } catch (ex) {
        console.log(ex);
        memphisConn && memphisConn.close();
    }
    })();`,
        consumer: `import memphis from "memphis-dev"

(async function () {
    let memphisConn;
    try {
        memphisConn = await memphis.connect({
            host: "<memphis-host>",
            username: "<application type username>",
            connectionToken: "<connection_token>"
        });
        
        const consumer = await memphisConn.consumer({
            stationName: "<station_name>",
            consumerName: "myConsumer",
            consumerGroup: ""
        });
        consumer.on("message", message => {
            console.log(message.getData().toString());
            message.ack();
        });
        consumer.on("error", error => {
            console.log(error);
        });
    } catch (ex) {
        console.log(ex);
        memphisConn && memphisConn.close();
    }
    })();`
    },

    Go: {
        langCode: 'go',
        installation: `go get github.com/memphisdev/memphis.go`,
        producer: `package main

import (
    "fmt"
    "os"
    "time"
    
    "github.com/memphisdev/memphis.go"
)
    
func main() {
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<connection_token>")
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()
    
    p, err := conn.CreateProducer("<station_name>", "myProducer")
    if err != nil {
        fmt.Printf("Producer creation failed: %v\n", err)
        os.Exit(1)
    }
    
    err = p.Produce([]byte("You have a message!"))
    if err != nil {
        fmt.Errorf("Produce failed: %v", err)
        os.Exit(1)
    }
}`,
        consumer: `package main

import (
    "fmt"
    "os"
    "time"
    
    "github.com/memphisdev/memphis.go"
)
    
func main() {
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<connection_token>")
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()
    
    consumer, err := conn.CreateConsumer("<station_name>", "myConsumer", memphis.PullInterval(15*time.Second))
    if err != nil {
        fmt.Printf("Consumer creation failed: %v\n", err)
        os.Exit(1)
    }
    
    handler := func(msgs []*memphis.Msg, err error) {
        if err != nil {
            fmt.Printf("Fetch failed: %v\n", err)
            return
        }
    
        for _, msg := range msgs {
            fmt.Println(string(msg.Data()))
            msg.Ack()
        }
    }
    
    consumer.Consume(handler)
    time.Sleep(30 * time.Second) //Keep your program running
    }`
    },

    Python: {
        langCode: 'python',
        installation: `pip3 install memphis-py`,
        producer: `import asyncio
from memphis import Memphis
        
async def main():
    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<connection_token>")
        
        producer = await memphis.producer(station_name="<station_name>", producer_name="myProducer")
        for i in range(100):
            await producer.produce(bytearray('Message #'+str(i)+': Hello world', 'utf-8'))
    except Exception as e:
        print(e)
    finally:
        await memphis.close()
        
if __name__ == '__main__':
    asyncio.run(main())`,
        consumer: `import asyncio
from memphis import Memphis
        
async def main():
    async def msg_handler(msgs, error):
        try:
            for msg in msgs:
                print(msg.get_data())
                await msg.ack()
            if error:
                print(error)
        except Exception as e:
            return

    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<connection_token>")
         
        consumer = await memphis.consumer(station_name="<station-name>", consumer_name="myConsumer", consumer_group="")
        consumer.consume(msg_handler)
        await asyncio.sleep(5)
    except Exception as e:
        print(e)
    finally:
        await memphis.close()
        
if __name__ == '__main__':
    asyncio.run(main())`
    }
};
