// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
export const sdkLangOptions = ['Go', 'Node.js', 'TypeScript', 'NestJS', 'Python', '.NET (C#)'];

export const SDK_CODE_EXAMPLE = {
    'Node.js': {
        langCode: 'javascript',
        installation: `npm i memphis-dev --save`,
        producer: `const { memphis } = require("memphis-dev");

(async function () {
    let memphisConnection

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>',
            accountId: "<account-id>"
        });

        const producer = await memphisConnection.producer({
            stationName: '<station-name>',
            producerName: '<producer-name>'
        });
        <headers-initiation>
        <headers-addition>

        await producer.produce({
            message: Buffer.from("Message: Hello world"), // you can also send JS object - {}
            headers: headers,
            <producer-async>
        });

        memphisConnection.close();
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `,
        consumer: `const { memphis } = require('memphis-dev');

(async function () {
    let memphisConnection;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>',
            accountId: "<account-id>"
        });

        const consumer = await memphisConnection.consumer({
            stationName: '<station-name>',
            consumerName: '<consumer-name>',
            consumerGroup: "<consumer-group>",
        });

        consumer.setContext({ key: "value" });
        consumer.on('message', (message, context) => {
            console.log(message.getData().toString());
            message.ack();
        });

        consumer.on('error', (error) => {});
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `
    },

    NestJS: {
        title: 'Please head over to the documentation',
        desc: "We'll provide you with snippets that you can easily connect your application with Memphis",
        link: 'https://github.com/memphisdev/memphis.js',
        installation: `npm i memphis-dev --save`
    },

    TypeScript: {
        langCode: 'typescript',
        installation: `npm i memphis-dev --save`,
        producer: `import { memphis, Memphis } from 'memphis-dev';

(async function () {
    let memphisConnection: Memphis | null = null;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>',
            accountId: "<account-id>"
        });

        const producer = await memphisConnection.producer({
            stationName: '<station-name>',
            producerName: '<producer-name>'
        });
            <headers-initiation>
            <headers-addition>

            await producer.produce({
                message: Buffer.from("Message: Hello world"), // you can also send JS object - {}
                headers: headers,
                <producer-async>
            });

        memphisConnection.close();
    } catch (ex) {
        console.log(ex);
        if (memphisConnection) memphisConnection.close();
    }
})();
        `,
        consumer: `import { memphis, Memphis, Message } from 'memphis-dev';

(async function () {
    let memphisConnection: Memphis | null = null;

    try {
        memphisConnection = await memphis.connect({
            host: '<memphis-host>',
            username: '<application type username>',
            connectionToken: '<broker-token>',
            accountId: "<account-id>"
        });

        const consumer = await memphisConnection.consumer({
            stationName: '<station-name>',
            consumerName: '<consumer-name>',
            consumerGroup: "<consumer-group>",
        });

        consumer.setContext({ key: "value" });
        consumer.on('message', (message: Message, context: object) => {
            console.log(message.getData().toString());
            message.ack();
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
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", memphis.ConnectionToken("<broker-token>"), memphis.AccountId("<account-id>"))
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()

    p, err := conn.CreateProducer("<station-name>", "<producer-name>")
    if err != nil {
        fmt.Printf("Producer failed: %v", err)
        os.Exit(1)
    }
    
    <headers-declaration>
    <headers-initiation>
    <headers-addition>

    err = p.Produce([]byte("You have a message!"), memphis.MsgHeaders(hdrs), <producer-async>)
    if err != nil {
        fmt.Printf("Produce failed: %v", err)
        os.Exit(1)
    }
}
        `,
        consumer: `package main

import (
    "fmt"
    "context"
    "os"
    "time"
    "github.com/memphisdev/memphis.go"
)

func main() {
    conn, err := memphis.Connect("<memphis-host>", "<application type username>", memphis.ConnectionToken("<broker-token>"), memphis.AccountId("<account-id>"))
    if err != nil {
        os.Exit(1)
    }
    defer conn.Close()

    consumer, err := conn.CreateConsumer("<station-name>", "<consumer-name>",memphis.ConsumerGroup("<consumer-group>"), memphis.PullInterval(15*time.Second))
    if err != nil {
        fmt.Printf("Consumer creation failed: %v", err)
        os.Exit(1)
    }

    handler := func(msgs []*memphis.Msg, err error, ctx context.Context) {
        if err != nil {
            fmt.Printf("Fetch failed: %v", err)
            return
        }

        for _, msg := range msgs {
            fmt.Println(string(msg.Data()))
            msg.Ack()
        }
    }

    ctx := context.Background()
	ctx = context.WithValue(ctx, "key", "value")
	consumer.SetContext(ctx)
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
        producer: `from __future__ import annotations
import asyncio
from memphis import Memphis, Headers, MemphisError, MemphisConnectError, MemphisHeaderError, MemphisSchemaError

async def main():
    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>", account_id="<account-id>")
        
        producer = await memphis.producer(station_name="<station-name>", producer_name="<producer-name>") # you can send the message parameter as dict as well
        
        <headers-initiation>
        <headers-addition>
    
        for i in range(5):
            await producer.produce(bytearray("Message #" + str(i) + ": Hello world", "utf-8"), headers=headers<blocking>)
        
    except (MemphisError, MemphisConnectError, MemphisHeaderError, MemphisSchemaError) as e:
        print(e)
        
    finally:
        await memphis.close()
        
if __name__ == "__main__":
    asyncio.run(main())`,
        consumer: `from __future__ import annotations
import asyncio
from memphis import Memphis, MemphisError, MemphisConnectError, MemphisHeaderError

async def main():
    async def msg_handler(msgs, error, context):
        try:
            for msg in msgs:
                print("message: ", msg.get_data())
                await msg.ack()
                if error:
                    print(error)
        except (MemphisError, MemphisConnectError, MemphisHeaderError) as e:
            print(e)
            return
        
    try:
        memphis = Memphis()
        await memphis.connect(host="<memphis-host>", username="<application type username>", connection_token="<broker-token>", account_id="<account-id>")
        
        consumer = await memphis.consumer(station_name="<station-name>", consumer_name="<consumer-name>", consumer_group="<consumer-group>")
        consumer.set_context({"key": "value"})
        consumer.consume(msg_handler)
        # Keep your main thread alive so the consumer will keep receiving data
        await asyncio.Event().wait()
        
    except (MemphisError, MemphisConnectError) as e:
        print(e)
        
    finally:
        await memphis.close()
        
if __name__ == "__main__":
    asyncio.run(main())`
    },

    '.NET (C#)': {
        langCode: 'C#',
        installation: `dotnet add package Memphis.Client`,
        producer: `using System.Collections.Specialized;
using System.Text;
using Memphis.Client;
using Memphis.Client.Producer;

namespace Producer
{
    class ProducerApp
    {
        public static async Task Main(string[] args)
        {
            try
            {
                var options = MemphisClientFactory.GetDefaultOptions();
                options.Host = "<memphis-host>";
                options.Username = "<application type username>";
                options.Password = "<password>";
                var client = await MemphisClientFactory.CreateClient(options);
                options.AccountId = "<account-id>";

                var producer = await client.CreateProducer(new MemphisProducerOptions
                {
                    StationName = "<station-name>",
                    ProducerName = "<producer-name>",
                    GenerateUniqueSuffix = true
                });

                <headers-declaration>
                
                


                <headers-initiation>
                <headers-addition>

                for (int i = 0; i < 10_000000; i++)
                {
                    await Task.Delay(1_000);
                    var text = $"Message #{i}: Welcome to Memphis";
                    await producer.ProduceAsync(Encoding.UTF8.GetBytes(text), commonHeaders, <producer-async>);
                    Console.WriteLine($"Message #{i} sent successfully");
                }

                await producer.DestroyAsync();
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine("Exception: " + ex.Message);
                Console.Error.WriteLine(ex);
            }
        }
    }
}
        `,
        consumer: `using System.Text;
using Memphis.Client;
using Memphis.Client.Consumer;

namespace Consumer
{
    class ConsumerApp
    {
        public static async Task Main(string[] args)
        {
            try
            {
                var options = MemphisClientFactory.GetDefaultOptions();
                options.Host = "<memphis-host>";
                options.Username = "<application type username>";
                options.Password = "<password>";
                var client = await MemphisClientFactory.CreateClient(options);
                options.AccountId = "<account-id>";

                var consumer = await client.CreateConsumer(new MemphisConsumerOptions
                {
                    StationName = "<station-name>",
                    ConsumerName = "<consumer-name>",
                    ConsumerGroup = "<consumer-group>",
                });

                consumer.MessageReceived += (sender, args) =>
                {
                    if (args.Exception != null)
                    {
                        Console.Error.WriteLine(args.Exception);
                        return;
                    }

                    foreach (var msg in args.MessageList)
                    {
                        //print message itself
                        Console.WriteLine("Received data: " + Encoding.UTF8.GetString(msg.GetData()));


                        // print message headers
                        foreach (var headerKey in msg.GetHeaders().Keys)
                        {
                            Console.WriteLine(
                                $"Header Key: {headerKey}, value: {msg.GetHeaders()[headerKey.ToString()]}");
                        }

                        Console.WriteLine("---------");
                        msg.Ack();
                    }
                    Console.WriteLine("destroyed");
                };

                consumer.ConsumeAsync();

                // Wait 10 seconds, consumer starts to consume, if you need block main thread use await keyword.
                await Task.Delay(10_000);
                await consumer.DestroyAsync();
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine("Exception: " + ex.Message);
                Console.Error.WriteLine(ex);
            }
        }
    }
}
        `
    }
};

export const restLangOptions = ['cURL', 'Go', 'Node.js', 'Python', 'Java', 'JavaScript - Fetch', 'JavaScript - jQuery'];
export const REST_CODE_EXAMPLE = {
    cURL: {
        langCode: 'apex',
        producer: `curl --location --request POST 'localhost/stations/<station-name>/produce/single' \\
--header 'Authorization: Bearer <jwt>' \\
--header 'Content-Type: application/json' \\
<headers-addition>
--data-raw '{"message": "New Message"}'`,
        consumer: `curl --location --request POST 'localhost/stations/<station-name>/consume/batch' \\
--header 'Authorization: Bearer <jwt>' \\
--header 'Content-Type: application/json' \\
--data-raw '{
    "consumer_name": "<consumer-name>",
    "consumer_group": "<consumer-group>",
    "batch_size": <batch-size>,
    "batch_max_wait_time_ms": <batch-max-wait-time-ms>\n}'`,
        tokenGenerate: `curl --location --request POST 'localhost/auth/authenticate' \\
--header 'Content-Type: application/json' \\
--data-raw '{
    "username": "<application type username>",
    "connection_token": "<broker-token>",
    "account_id": "<account-id>",
    "token_expiry_in_minutes": <token_expiry_in_minutes>,
    "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>\n}'`
    },
    Go: {
        langCode: 'go',
        producer: `package main
    import (
      "fmt"
      "strings"
      "net/http"
      "io"
    )
    
    func main() {
    
        url := "localhost/stations/<station-name>/produce/single"
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
        <headers-addition>

        res, err := client.Do(req)
        if err != nil {
            fmt.Println(err)
            return
        }
        defer res.Body.Close()
      
        body, err := io.ReadAll(res.Body)
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(string(body))
    }`,
        consumer: `package main
    import (
      "fmt"
      "strings"
      "net/http"
      "io"
    )
    
    func main() {
    
        url := "localhost/stations/<station-name>/consume/batch"
        method := "POST"
      
        payload := strings.NewReader('{
            "consumer_name": "<consumer-name>",
            "consumer_group": "<consumer-group>",
            "batch_size": <batch-size>,
            "batch_max_wait_time_ms": <batch-max-wait-time-ms>
        })
      
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
      
        body, err := io.ReadAll(res.Body)
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
      "io"
    )
    
    func main() {
    
        url := "localhost/auth/authenticate"
        method := "POST"
      
        payload := strings.NewReader('{
          "username": "<application type username>",
          "connection_token": "<broker-token>",
          "account_id": "<account-id>",
          "token_expiry_in_minutes": <token_expiry_in_minutes>,
          "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
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
      
        body, err := io.ReadAll(res.Body)
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
  url: 'localhost/stations/<station-name>/produce/single',
  headers: { 
    'Authorization': 'Bearer <jwt>', 
    'Content-Type': 'application/json',
    <headers-addition>
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
        consumer: `var axios = require('axios');
        var data = JSON.stringify({
            "consumer_name": "<consumer-name>",
            "consumer_group": "<consumer-group>",
            "batch_size": <batch-size>,
            "batch_max_wait_time_ms": <batch-max-wait-time-ms>
          });

var config = {
  method: 'post',
  url: 'localhost/stations/<station-name>/consume/batch',
  headers: {
    'Authorization': 'Bearer <jwt>', 
    'Content-Type': 'application/json',
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
  "username": "<application type username>",
  "connection_token": "<broker-token>",
  "account_id": "<account-id>",
  "token_expiry_in_minutes": <token_expiry_in_minutes>,
  "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
});

var config = {
  method: 'post',
  url: 'localhost/auth/authenticate',
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

url = "localhost/stations/<station-name>/produce/single"

payload = json.dumps({
  "message": "New Message"
})
headers = {
  'Authorization': 'Bearer <jwt>',
  'Content-Type': 'application/json',
  <headers-addition>
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
`,
        consumer: `import requests
import json

url = "localhost/stations/<station-name>/consume/batch"

payload = json.dumps({
    "consumer_name": "<consumer-name>",
    "consumer_group": "<consumer-group>",
    "batch_size": <batch-size>,
    "batch_max_wait_time_ms": <batch-max-wait-time-ms>
  })
headers = {
  'Authorization': 'Bearer <jwt>',
  'Content-Type': 'application/json',
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
`,
        tokenGenerate: `import requests
import json

url = "localhost/auth/authenticate"

payload = json.dumps({
  "username": "<application type username>",
  "connection_token": "<broker-token>",
  "account_id": "<account-id>",
  "token_expiry_in_minutes": <token_expiry_in_minutes>,
  "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
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
  .url("localhost/stations/<station-name>/produce/single")
  .method("POST", body)
  .addHeader("Authorization", "Bearer <jwt>")
  .addHeader("Content-Type", "application/json")
  <headers-addition>
  .build();
Response response = client.newCall(request).execute();`,
        consumer: `OkHttpClient client = new OkHttpClient().newBuilder()
.build();
MediaType mediaType = MediaType.parse("application/json");
RequestBody body = RequestBody.create(mediaType, "{\n    \"consumer_name\": \"<consumer-name>\",\n\t\"consumer_group\": \"<consumer-group>\",\n    \"batch_size\": <batch-size>,\n    \"batch_max_wait_time_ms\": <batch-max-wait-time-ms>\n}");
Request request = new Request.Builder()
.url("localhost/stations/<station-name>/consume/batch")
.method("POST", body)
.addHeader("Authorization", "Bearer <jwt>")
.addHeader("Content-Type", "application/json")
<headers-addition>
.build();
Response response = client.newCall(request).execute();`,
        tokenGenerate: `OkHttpClient client = new OkHttpClient().newBuilder()
  .build();
MediaType mediaType = MediaType.parse("application/json");
RequestBody body = RequestBody.create(mediaType, "{\n    \"username\": \"<application type username>\",\n\t\"connection_token\": \"<broker-token>\",\n    \"token_expiry_in_minutes\": <token_expiry_in_minutes>,\n    \"refresh_token_expiry_in_minutes\": <refresh_token_expiry_in_minutes>\n    \"account_id\": \"<account-id>\"\n}");
Request request = new Request.Builder()
  .url("localhost/auth/authenticate")
  .method("POST", body)
  .addHeader("Content-Type", "application/json")
  .build();
Response response = client.newCall(request).execute();`
    },
    'JavaScript - Fetch': {
        langCode: 'javascript',
        producer: `const fetch = require('node-fetch');
const myHeaders = new fetch.Headers();
myHeaders.append("Authorization", "Bearer <jwt>");
myHeaders.append("Content-Type", "application/json");
<headers-addition>

var raw = JSON.stringify({
  "message": "New Message"
});

var requestOptions = {
  method: 'POST',
  headers: myHeaders,
  body: raw,
  redirect: 'follow'
};

fetch("localhost/stations/<station-name>/produce/single", requestOptions)
  .then(response => response.text())
  .then(result => console.log(result))
  .catch(error => console.log('error', error));`,
        consumer: `const fetch = require('node-fetch');
  const myHeaders = new fetch.Headers();
  myHeaders.append("Authorization", "Bearer <jwt>");
  myHeaders.append("Content-Type", "application/json");
  
  var raw = JSON.stringify({
    "consumer_name": "<consumer-name>",
    "consumer_group": "<consumer-group>",
    "batch_size": <batch-size>,
    "batch_max_wait_time_ms": <batch-max-wait-time-ms>
  });
  
  var requestOptions = {
    method: 'POST',
    headers: myHeaders,
    body: raw,
    redirect: 'follow'
  };
  
  fetch("localhost/stations/<station-name>/consume/batch", requestOptions)
    .then(response => response.text())
    .then(result => console.log(result))
    .catch(error => console.log('error', error));`,
        tokenGenerate: `const fetch = require('node-fetch');
const myHeaders = new fetch.Headers();
myHeaders.append("Content-Type", "application/json");

var raw = JSON.stringify({
  "username": "<application type username>",
  "connection_token": "<broker-token>",
  "account_id": "<account-id>",
  "token_expiry_in_minutes": <token_expiry_in_minutes>,
  "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
});

var requestOptions = {
  method: 'POST',
  headers: myHeaders,
  body: raw,
  redirect: 'follow'
};

fetch("localhost/auth/authenticate", requestOptions)
  .then(response => response.text())
  .then(result => console.log(result))
  .catch(error => console.log('error', error));`
    },
    'JavaScript - jQuery': {
        langCode: 'javascript',
        producer: `var settings = {
  "url": "localhost/stations/<station-name>/produce/single",
  "method": "POST",
  "timeout": 0,
  "headers": {
    "Authorization": "Bearer <jwt>",
    "Content-Type": "application/json",
    <headers-addition>
  "data": JSON.stringify({
    "message": "New Message"
  }),
};

$.ajax(settings).done(function (response) {
console.log(response);
});`,
        consumer: `var settings = {
    "url": "localhost/stations/<station-name>/consume/batch",
    "method": "POST",
    "timeout": 0,
    "headers": {
      "Authorization": "Bearer <jwt>",
      "Content-Type": "application/json",
    "data": JSON.stringify({
        "consumer_name": "<consumer-name>",
        "consumer_group": "<consumer-group>",
        "batch_size": <batch-size>,
        "batch_max_wait_time_ms": <batch-max-wait-time-ms>
    }),
  };
  
  $.ajax(settings).done(function (response) {
  console.log(response);
  });`,
        tokenGenerate: `var settings = {
  "url": "localhost/auth/authenticate",
  "method": "POST",
  "timeout": 0,
  "headers": {
    "Content-Type": "application/json"
  },
  "data": JSON.stringify({
    "username": "<application type username>",
    "connection_token": "<broker-token>",
    "account_id": "<account-id>",
    "token_expiry_in_minutes": <token_expiry_in_minutes>,
    "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
  }),
  };

$.ajax(settings).done(function (response) {
console.log(response);
});`
    },
    '.NET (C#)': {
        langCode: 'C#',
        producer: `using System;
using System.Net.Http;
using System.Threading.Tasks;

internal class Program
{
    private static async Task Main(string[] args)
    {
        var client = new HttpClient();
        var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost:4444/stations/test/produce/single");
        request.Headers.Add("Authorization", "Bearer <jwt>");
        var content = new StringContent("""
        {
            "message": "New Message"
        }
        """, null, "application/json");
        request.Content = content;
        var response = await client.SendAsync(request);
        response.EnsureSuccessStatusCode();
        Console.WriteLine(await response.Content.ReadAsStringAsync());
    }
}
        `,
        consumer: `using System;
        using System.Net.Http;
        using System.Threading.Tasks;
        
        internal class Program
        {
            private static async Task Main(string[] args)
            {
                var client = new HttpClient();
                var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost:4444/stations/test/consume/batch");
                request.Headers.Add("Authorization", "Bearer <jwt>");
                var content = new StringContent("""
                { 
                    "consumer_name": "<consumer-name>",
                    "consumer_group": "<consumer-group>",
                    "batch_size": <batch-size>,
                    "batch_max_wait_time_ms": <batch-max-wait-time-ms>
                }
                """, null, "application/json");
                request.Content = content;
                var response = await client.SendAsync(request);
                response.EnsureSuccessStatusCode();
                Console.WriteLine(await response.Content.ReadAsStringAsync());
            }
        }
                `,
        tokenGenerate: `using System;
using System.Net.Http;
using System.Threading.Tasks;

internal class Program
{
    private static async Task Main(string[] args)
    {
        var client = new HttpClient();
        var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost:4444/auth/authenticate");
        var content = new StringContent("""
        { 
            "username": "<application type username>",    
            "password": "<password>",
            "token_expiry_in_minutes": <token_expiry_in_minutes>,
            "refresh_token_expiry_in_minutes": <refresh_token_expiry_in_minutes>
        }
        """, null, "application/json");
        request.Content = content;
        var response = await client.SendAsync(request);
        response.EnsureSuccessStatusCode();
        Console.WriteLine(await response.Content.ReadAsStringAsync());
    }
}
        `
    }
};
