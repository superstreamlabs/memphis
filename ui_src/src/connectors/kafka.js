export const kafka = {
    Source: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap.servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'kafka-1:9092,kafka-2:9092'
        },
        {
            name: 'security.protocol',
            display: 'Security protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl.mechanism',
                    display: 'SSL mechanism',
                    type: 'select',
                    options: ['GSSAPI', 'PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512'],
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name: 'ssl.key.password',
                    display: 'SSL key password',
                    type: 'string',
                    required: true
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl.mechanism',
                    display: 'SASL mechanism',
                    type: 'select',
                    options: ['GSSAPI', 'PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512'],
                    required: true
                },
                {
                    name: 'sasl.username',
                    display: 'SASL username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl.password',
                    display: 'SASL password',
                    type: 'string',
                    required: true,
                    placeholder: 'password'
                }
            ],
            'No authentication': []
        },
        {
            name: 'group.id',
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
            name: 'partition_strategy',
            display: 'Partition strategy',
            type: 'select',
            options: ['Partition Number', 'Any Partition'],
            required: true,
            description: 'Partition Number / Any Partition (round robin)',
            children: true,
            'Partition Number': [
                {
                    name: 'partition_value',
                    display: 'Partition Value',
                    display: 'Partition Value',
                    type: 'string',
                    required: true
                },
                {
                    name: 'offset_strategy',
                    display: 'Offset strategy',
                    type: 'select',
                    options: ['Earliest', 'End', 'Specific offset'],
                    required: false,
                    description: 'choose offset strategy',
                    children: true,
                    Earliest: [],
                    End: [],
                    'Specific offset': [
                        {
                            name: 'offset_value',
                            description: 'choose offset value (int)',
                            display: 'Value',
                            type: 'string',
                            required: true,
                            placeholder: 0
                        }
                    ]
                }
            ],
            'Any Partition': [
                {
                    name: 'offset_strategy',
                    display: 'Offset strategy',
                    type: 'select',
                    options: ['Earliest', 'End'],
                    required: false,
                    description: 'choose offset strategy'
                }
            ]
            'Any Partition': [
                {
                    name: 'offset_strategy',
                    display: 'Offset strategy',
                    type: 'select',
                    options: ['Earliest', 'End'],
                    required: false,
                    description: 'choose offset strategy'
                }
            ]
        },
        {
            name: 'timeout_duration_seconds',
            display: 'Consumer timeout duration (seconds)',
            type: 'string',
            required: false,
            placeholder: 10
        }
    ],
    Sink: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true,
            description: 'Note that the sink connector name is also consumer group name'
        },
        {
            name: 'bootstrap.servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'kafka-1:9092,kafka-2:9092'
        },
        {
            name: 'security.protocol',
            display: 'Security protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl.mechanism',
                    display: 'SSL mechanism',
                    type: 'select',
                    options: ['GSSAPI', 'PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512'],
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true,
                    placeholder: '-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----'
                },
                {
                    name: 'ssl.key.password',
                    display: 'SSL key password',
                    type: 'string',
                    required: true
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl.mechanism',
                    display: 'SASL mechanism',
                    type: 'select',
                    options: ['GSSAPI', 'PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512'],
                    required: true
                },
                {
                    name: 'sasl.username',
                    display: 'SASL username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl.password',
                    display: 'SASL password',
                    type: 'string',
                    required: true
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
            options: ['Partition Key', 'Partition Number', 'Any Partition'],
            required: true,
            description: 'Partition Key / Partition Number / Any Partition',
            children: true,
            'Partition Key': [
                {
                    name: 'partition_value',
                    display: 'Value',
                    type: 'string',
                    required: true
                }
            ],
            'Partition Number': [
                {
                    name: 'partition_value',
                    display: 'Value',
                    type: 'string',
                    required: true
                }
            ],
            'Any Partition': []
        },
        {
            name: 'memphis_batch_size',
            display: 'Memphis batch size (messages)',
            type: 'string',
            required: false,
            placeholder: 100,
            description: 'The buffer size used by Memphis to accumulate and handle incoming messages before processing'
        },
        {
            name: 'memphis_max_time_wait',
            display: 'Max time to wait for a batch of messages (seconds)',
            placeholder: 5,
            type: 'string',
            required: false,
            description: 'The duration which a batch of messages is awaited till processing'
        }
    ]
};
