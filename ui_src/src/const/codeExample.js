// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

export const SDK_CODE_EXAMPLE = {
    'Node.js': {
        langCode: 'javascript',
        installation: `npm i memphis-dev --save`,
        producer: `const memphis = require("memphis-dev");

(async function () {
    let memphisConnection

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>'
        });

        const producer = await memphisConnection.producer({
            stationName: '<station-name>',
            producerName: '<producer-name>'
        });

        const headers = memphis.headers()
        headers.add('key', 'value')
        await producer.produce({
            message: Buffer.from("Message: Hello world"),
            headers: headers
        });

        memphisConnection.close();
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `,
        consumer: `const memphis = require('memphis-dev');

(async function () {
    let memphisConnection;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>'
        });

        const consumer = await memphisConnection.consumer({
            stationName: '<station-name>',
            consumerName: '<consumer-name>',
            consumerGroup: ''
        });

        consumer.on('message', (message) => {
            console.log(message.getData().toString());
            message.ack();
            const headers = message.getHeaders()
        });

        consumer.on('error', (error) => {});
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `
    },

    TypeScript: {
        langCode: 'typescript',
        installation: `npm i memphis-dev --save`,
        producer: `import memphis from 'memphis-dev';
import type { Memphis } from 'memphis-dev/types';

(async function () {
    let memphisConnection: Memphis;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>'
        });

        const producer = await memphisConnection.producer({
            stationName: '<station-name>',
            producerName: '<producer-name>'
        });

            const headers = memphis.headers()
            headers.add('key', 'value');
            await producer.produce({
                message: Buffer.from("Message: Hello world"),
                headers: headers
            });

        memphisConnection.close();
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `,
        consumer: `import memphis from 'memphis-dev';
import { Memphis, Message } from 'memphis-dev/types';

(async function () {
    let memphisConnection: Memphis;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>'
        });

        const consumer = await memphisConnection.consumer({
            stationName: '<station-name>',
            consumerName: '<consumer-name>',
            consumerGroup: ''
        });

        consumer.on('message', (message: Message) => {
            console.log(message.getData().toString());
            message.ack();
            const headers = message.getHeaders()
        });

        consumer.on('error', (error) => {
            console.log(error);
        });
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `
    },

    Go: {
        langCode: 'go',
        installation: `go get github.com/memphisdev/memphis.go`,
        producer: `package main

import (
    "fmt"
    "os"

    "github.com/memphisdev/memphis.go"
)

func main() {
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<broker-token>")
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()
    p, err := conn.CreateProducer("<station-name>", "<producer-name>")

    hdrs := memphis.Headers{}
    hdrs.New()
    err = hdrs.Add("key", "value")

    if err != nil {
        fmt.Printf("Header failed: %v\n", err)
        os.Exit(1)
    }

    err = p.Produce([]byte("You have a message!"), memphis.MsgHeaders(hdrs))

    if err != nil {
        fmt.Printf("Produce failed: %v\n", err)
        os.Exit(1)
    }
}
        `,
        consumer: `package main

import (
    "fmt"
    "os"
    "time"

    "github.com/memphisdev/memphis.go"
)

func main() {
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", "<broker-token>")
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()

    consumer, err := conn.CreateConsumer("<station-name>", "<consumer-name>", memphis.PullInterval(15*time.Second))

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
            headers := msg.GetHeaders()
            fmt.Println(headers)
        }
    }

    consumer.Consume(handler)

    // The program will close the connection after 30 seconds,
    // the message handler may be called after the connection closed
    // so the handler may receive a timeout error
    time.Sleep(30 * time.Second)
}
`
    },

    Python: {
        langCode: 'python',
        installation: `pip3 install --upgrade memphis-py`,
        producer: `import asyncio
from memphis import Memphis, Headers, MemphisError, MemphisConnectError, MemphisHeaderError, MemphisSchemaError
        
async def main():
    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>")
        
        producer = await memphis.producer(station_name="<station-name>", producer_name="<producer-name>")
        headers = Headers()
        headers.add("key", "value") 
        for i in range(5):
            await producer.produce(bytearray('Message #'+str(i)+': Hello world', 'utf-8'), headers=headers)
        
    except (MemphisError, MemphisConnectError, MemphisHeaderError, MemphisSchemaError) as e:
        print(e)
        
    finally:
        await memphis.close()
        
if __name__ == '__main__':
    asyncio.run(main())
        `,
        consumer: `import asyncio
from memphis import Memphis, MemphisError, MemphisConnectError, MemphisHeaderError
        
async def main():
    async def msg_handler(msgs, error):
        try:
            for msg in msgs:
                print("message: ", msg.get_data())
                await msg.ack()
                headers = msg.get_headers()
                if error:
                    print(error)
        except (MemphisError, MemphisConnectError, MemphisHeaderError) as e:
            print(e)
            return
        
    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>")
        
        consumer = await memphis.consumer(station_name="<station-name>", consumer_name="<consumer-name>", consumer_group="")
        consumer.consume(msg_handler)
        # Keep your main thread alive so the consumer will keep receiving data
        await asyncio.Event().wait()
        
    except (MemphisError, MemphisConnectError) as e:
        print(e)
        
    finally:
        await memphis.close()
        
if __name__ == '__main__':
    asyncio.run(main())
        `
    }
};

export const PROTOCOL_CODE_EXAMPLE = {
    cURL: {
        langCode: 'apex',
        producer: `curl --location --request POST 'localhost:4444/stations/<station-name>/produce/single' \\\n--header 'Authorization: Bearer <jwt>' \\\n--header 'Content-Type: application/json' \\\n--data-raw '{"message": "New Message"}'`,
        tokenGenerate: `curl --location --request POST 'localhost:4444/auth/authenticate' \\\n--header 'Content-Type: application/json' \\\n--data-raw '{
    "username": "root",
    "connection_token": "memphis",
    "token_expiry_in_minutes": 123,
    "refresh_token_expiry_in_minutes": 10000092\n}'`
    },
    Go: {
        langCode: 'go',
        producer: `package main

        import (
          "fmt"
          "strings"
          "net/http"
          "io/ioutil"
        )
        
        func main() {
        
          url := "localhost:4444/stations/<station-name>/produce/single"
          method := "POST"
        
          payload := strings.NewReader('{"message": "New Message"}')
        
          client := &http.Client {
          }
          req, err := http.NewRequest(method, url, payload)
        
          if err != nil {
            fmt.Println(err)
            return
          }
          req.Header.Add("Authorization", "Bearer <jwt>")
          req.Header.Add("Content-Type", "application/json")
        
          res, err := client.Do(req)
          if err != nil {
            fmt.Println(err)
            return
          }
          defer res.Body.Close()
        
          body, err := ioutil.ReadAll(res.Body)
          if err != nil {
            fmt.Println(err)
            return
          }
          fmt.Println(string(body))
        }`,
        tokenGenerate: `package main

        import (
          "fmt"
          "strings"
          "net/http"
          "io/ioutil"
        )
        
        func main() {
        
          url := "localhost:4444/auth/authenticate"
          method := "POST"
        
          payload := strings.NewReader({
            "username": "root",
            "connection_token": "memphis",
            "token_expiry_in_minutes": 123,
            "refresh_token_expiry_in_minutes": 10000092
        })
        
          client := &http.Client {
          }
          req, err := http.NewRequest(method, url, payload)
        
          if err != nil {
            fmt.Println(err)
            return
          }
          req.Header.Add("Content-Type", "application/json")
        
          res, err := client.Do(req)
          if err != nil {
            fmt.Println(err)
            return
          }
          defer res.Body.Close()
        
          body, err := ioutil.ReadAll(res.Body)
          if err != nil {
            fmt.Println(err)
            return
          }
          fmt.Println(string(body))
        }`
    },
    'Node.js': {
        langCode: 'javascript',
        producer: `var axios = require('axios');
        var data = JSON.stringify({
          "message": "New Message"
        });
        
        var config = {
          method: 'post',
          url: 'localhost:4444/stations/<station-name>/produce/single',
          headers: { 
            'Authorization': 'Bearer <jwt>', 
            'Content-Type': 'application/json'
          },
          data : data
        };
        
        axios(config)
        .then(function (response) {
          console.log(JSON.stringify(response.data));
        })
        .catch(function (error) {
          console.log(error);
        });
        `,
        tokenGenerate: `var axios = require('axios');
        var data = JSON.stringify({
          "username": "root",
          "connection_token": "memphis",
          "token_expiry_in_minutes": 123,
          "refresh_token_expiry_in_minutes": 10000092
        });
        
        var config = {
          method: 'post',
          url: 'localhost:4444/auth/authenticate',
          headers: { 
            'Content-Type': 'application/json'
          },
          data : data
        };
        
        axios(config)
        .then(function (response) {
          console.log(JSON.stringify(response.data));
        })
        .catch(function (error) {
          console.log(error);
        });
        `
    },
    Python: {
        langCode: 'python',
        producer: `import requests
import json

url = "localhost:4444/stations/<station-name>/produce/single"

payload = json.dumps({
  "message": "New Message"
})
headers = {
  'Authorization': 'Bearer <jwt>',
  'Content-Type': 'application/json'
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
`,
        tokenGenerate: `import requests
import json

url = "localhost:4444/auth/authenticate"

payload = json.dumps({
  "username": "root",
  "connection_token": "memphis",
  "token_expiry_in_minutes": 123,
  "refresh_token_expiry_in_minutes": 10000092
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
        `
    },
    Java: {
        langCode: 'java',
        producer: `OkHttpClient client = new OkHttpClient().newBuilder()
  .build();
MediaType mediaType = MediaType.parse("application/json");
RequestBody body = RequestBody.create(mediaType, "{\"message\": \"New Message\"}");
Request request = new Request.Builder()
  .url("localhost:4444/stations/<station-name>/produce/single")
  .method("POST", body)
  .addHeader("Authorization", "Bearer <jwt>")
  .addHeader("Content-Type", "application/json")
  .build();
Response response = client.newCall(request).execute();`,
        tokenGenerate: `OkHttpClient client = new OkHttpClient().newBuilder()
  .build();
MediaType mediaType = MediaType.parse("application/json");
RequestBody body = RequestBody.create(mediaType, "{\n    \"username\": \"root\",\n\t\"connection_token\": \"memphis\",\n    \"token_expiry_in_minutes\": 123,\n    \"refresh_token_expiry_in_minutes\": 10000092\n}");
Request request = new Request.Builder()
  .url("localhost:4444/auth/authenticate")
  .method("POST", body)
  .addHeader("Content-Type", "application/json")
  .build();
Response response = client.newCall(request).execute();`
    },
    'JavaScript - Fetch': {
        langCode: 'javascript',
        producer: `var myHeaders = new Headers();
        myHeaders.append("Authorization", "Bearer <jwt>");
        myHeaders.append("Content-Type", "application/json");
        
        var raw = JSON.stringify({
          "message": "New Message"
        });
        
        var requestOptions = {
          method: 'POST',
          headers: myHeaders,
          body: raw,
          redirect: 'follow'
        };
        
        fetch("localhost:4444/stations/<station-name>/produce/single", requestOptions)
          .then(response => response.text())
          .then(result => console.log(result))
          .catch(error => console.log('error', error));`,
        tokenGenerate: `var myHeaders = new Headers();
        myHeaders.append("Content-Type", "application/json");
        
        var raw = JSON.stringify({
          "username": "root",
          "connection_token": "memphis",
          "token_expiry_in_minutes": 123,
          "refresh_token_expiry_in_minutes": 10000092
        });
        
        var requestOptions = {
          method: 'POST',
          headers: myHeaders,
          body: raw,
          redirect: 'follow'
        };
        
        fetch("localhost:4444/auth/authenticate", requestOptions)
          .then(response => response.text())
          .then(result => console.log(result))
          .catch(error => console.log('error', error));`
    },
    'JavaScript - jQuery': {
        langCode: 'javascript',
        producer: `var settings = {
            "url": "localhost:4444/stations/<station-name>/produce/single",
            "method": "POST",
            "timeout": 0,
            "headers": {
              "Authorization": "Bearer <jwt>",
              "Content-Type": "application/json"
            },
            "data": JSON.stringify({
              "message": "New Message"
            }),
          };
          
          $.ajax(settings).done(function (response) {
            console.log(response);
          });`,
        tokenGenerate: `var settings = {
            "url": "localhost:4444/auth/authenticate",
            "method": "POST",
            "timeout": 0,
            "headers": {
              "Content-Type": "application/json"
            },
            "data": JSON.stringify({
              "username": "root",
              "connection_token": "memphis",
              "token_expiry_in_minutes": 123,
              "refresh_token_expiry_in_minutes": 10000092
            }),
          };
          
          $.ajax(settings).done(function (response) {
            console.log(response);
          });`
    }
};
