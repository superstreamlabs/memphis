export const kafka = {
    Source: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap_servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'kafka-1:9092,kafka-2:9092'
        },
        {
            name: 'security_protocol',
            display: 'Security protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl_key_pem',
                    display: 'SSL Key Pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN PRIVATE KEY----- \n...\n-----END PRIVATE KEY-----'
                },
                {
                    name: 'ssl_certificate_pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name: 'ssl_ca_pem',
                    display: 'SSL CA PEM',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name:'insecure_skip_verify',
                    display:'Insecure skip verify',
                    type:'select',
                    options:['true','false'],
                    required:true,
                    description:'true / false'
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl_mechanism',
                    display: 'SASL mechanism',
                    type: 'select',
                    options: ['SCRAM-SHA-256', 'SCRAM-SHA-512', 'PLAIN'],
                    required: true
                },
                {
                    name: 'sasl_username',
                    display: 'SASL username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl_password',
                    display: 'SASL password',
                    type: 'string',
                    required: true,
                    placeholder: 'password'
                },
                {
                    name: 'tls_enabled',
                    display: 'TLS enabled',
                    type: 'select',
                    options: ['custom', 'default', 'none'],
                    required: true,
                    description: 'custom / default / none (no tls)',
                    children: true,
                    custom: [
                        {
                            name: 'ssl_key_pem',
                            display: 'SSL Key Pem',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN PRIVATE KEY----- \n...\n-----END PRIVATE KEY-----'
                        },
                        {
                            name: 'ssl_certificate_pem',
                            display: 'SSL certificate pem',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                        },
                        {
                            name: 'ssl_ca_pem',
                            display: 'SSL CA PEM',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                        },
                        {
                            name:'insecure_skip_verify',
                            display:'Insecure skip verify',
                            type:'select',
                            options:['true','false'],
                            required:true,
                            description:'true / false'
                        }
                    ],
                    default: [],
                    none: []
                }
            ],
            'No authentication': []
        },
        {
            name: 'group_id',
            display: 'Group id',
            type: 'string',
            required: true,
            description: 'Consumer group id'
        },
        {
            name: 'topic',
            display: 'Topic',
            type: 'string',
            required: true,
            description: 'Topic name'
        },
        {
            name: 'offset_strategy',
            display: 'Offset strategy',
            type: 'select',
            options: ['Newest', 'Oldest'],
            required: true,
            description: 'Newest / Oldest',
        },
        {
            name: 'fetch_size_bytes',
            display: 'Fetch size (bytes)',
            type: 'string',
            required: false,
            placeholder: 1000,
            description: 'The buffer size used by Kafka Consumer (in bytes)'
        },
        {
            name: 'fetch_max_wait_ms',
            display: 'Fetch Timeout Duration (Milliseconds)',
            placeholder: 1,
            type: 'string',
            required: false,
            description: 'The wait time before fetching the buffer (in milliseconds)',
        },
        {
            name: 'instances',
            display: 'Scale (instances)',
            placeholder: 1,
            min: 1,
            max: 15,
            type: 'number',
            required: false,
            description: 'The number of the connector instances '
        }
    ],
    Sink: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap_servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'kafka-1:9092,kafka-2:9092'
        },
        {
            name: 'security_protocol',
            display: 'Security protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl_key_pem',
                    display: 'SSL Key Pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN PRIVATE KEY----- \n...\n-----END PRIVATE KEY-----'
                },
                {
                    name: 'ssl_certificate_pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name: 'ssl_ca_pem',
                    display: 'SSL CA PEM',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name:'insecure_skip_verify',
                    display:'Insecure skip verify',
                    type:'select',
                    options:['true','false'],
                    required:true,
                    description:'true / false'
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl_mechanism',
                    display: 'SASL mechanism',
                    type: 'select',
                    options: ['SCRAM-SHA-256', 'SCRAM-SHA-512', 'PLAIN'],
                    required: true
                },
                {
                    name: 'sasl_username',
                    display: 'SASL username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl_password',
                    display: 'SASL password',
                    type: 'string',
                    required: true,
                    placeholder: 'password'
                },
                {
                    name: 'tls_enabled',
                    display: 'TLS enabled',
                    type: 'select',
                    options: ['custom', 'default', 'none'],
                    required: true,
                    description: 'custom / default / none (no tls)',
                    children: true,
                    true: [
                        {
                            name: 'ssl_key_pem',
                            display: 'SSL Key Pem',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN PRIVATE KEY----- \n...\n-----END PRIVATE KEY-----'
                        },
                        {
                            name: 'ssl_certificate_pem',
                            display: 'SSL certificate pem',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                        },
                        {
                            name: 'ssl_ca_pem',
                            display: 'SSL CA PEM',
                            type: 'string',
                            required: true,
                            placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                        },
                        {
                            name:'insecure_skip_verify',
                            display:'Insecure skip verify',
                            type:'select',
                            options:['true','false'],
                            required:true,
                            description:'true / false'
                        }
                    ],
                    false: [],
                    none: []
                }
            ],
            'No authentication': []
        },
        {
            name: 'topic',
            display: 'Topic',
            type: 'string',
            required: true,
            description: 'Topic name'
        },
        {
            name: 'partition_strategy',
            display: 'Partition strategy',
            type: 'select',
            options: ['Partition Number', 'Partition Key', 'Any Partition'],
            required: true,
            description: 'Partition Number / Partition Key / Any Partition (round robin)',
            children: true,
            'Partition Number': [
                {
                    name: 'partition_value',
                    display: 'Partition Value',
                    type: 'string',
                    required: true
                }
            ],
            'Partition Key': [
                {
                    name: 'partition_value',
                    display: 'Partition Key',
                    type: 'string',
                    required: true
                }
            ],
            'Any Partition': []
        },
        {
            name: 'flush_msg_number',
            display: 'Flush Message Number',
            type: 'string',
            required: false,
            placeholder: 100,
            description: 'The buffer size used by Kafka Producer (in messages)'
        },
        {
            name: 'flush_frequency',
            display: 'Flush Timeout Duration (Milliseconds)',
            placeholder: 1,
            type: 'string',
            required: false,
            description: 'The wait time before flushing the buffer (in milliseconds)'
        },
        {
            name: 'consume_from',
            display: 'Start consume from the beginning / end',
            type: 'select',
            options: ['Beginning', 'End'],
            description: 'Beginning (oldest messages) / End (newest messages) of the station',
        },
        {
            name: 'memphis_batch_size',
            display: 'Batch size (messages)',
            type: 'string',
            required: false,
            placeholder: 100,
            description: 'The buffer size used by Memphis to accumulate and handle incoming messages before processing'
        },
        {
            name: 'memphis_max_time_wait',
            display: 'Batch Message Timeout Duration (Milliseconds)',
            placeholder: 2,
            type: 'string',
            required: false,
            description: 'The wait time before delivering a batch of messages'
        },
        {
            name: 'instances',
            display: 'Scale (instances)',
            placeholder: 1,
            min: 1,
            max: 15,
            type: 'number',
            required: false,
            description: 'The number of the connector instances '
        }
    ]
};
