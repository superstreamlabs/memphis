export const kafka = {
    Source: [
        {
            name: 'name',
            display: 'Name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap.servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'localhost:9092',
            description: 'list of brokers'
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
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true
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
                    options: ['PLAIN', 'SCRAM-SHA-256'],
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
            name: 'group.id',
            display: 'Group id',
            type: 'string',
            required: true,
            description: 'consumer group id'
        },
        {
            name: 'offset_strategy',
            display: 'Offset strategy',
            type: 'select',
            options: ['Earliest', 'End', 'Specific offset (int)'],
            required: false,
            description: 'choose offset strategy',
            children: true,
            Earliest: [],
            End: [],
            'Specific offset (int)': [
                {
                    name: 'offset_value',
                    display: 'Value',
                    type: 'string',
                    required: true
                }
            ]
        },
        {
            name: 'topic',
            display: 'Topic',
            type: 'string',
            required: true,
            description: 'topic name'
        },
        {
            name: 'partition_strategy',
            display: 'Partition strategy',
            type: 'select',
            options: ['Partition Number', 'Any Partition'],
            required: true,
            description: 'Partition Number / Any Partition',
            children: true,
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
            name: 'timeout_duration_seconds',
            display: 'kafka consumer timeout duration',
            type: 'string',
            required: false,
            description: 'kafka consumer timeout duration'
        }
    ],
    Sink: [
        {
            name: 'name',
            display: 'Name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap.servers',
            display: 'Bootstrap servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'localhost:9092',
            description: 'list of brokers'
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
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'SSL certificate pem',
                    type: 'string',
                    required: true
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
                    options: ['PLAIN', 'SCRAM-SHA-256', 'SCRAM-SHA-512'],
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
            description: 'topic name'
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
            display: 'Memphis batch size',
            type: 'string',
            required: false,
            description: 'memphis consuemr batch size'
        },
        {
            name: 'memphis_max_time_wait',
            display: 'Max time to wait for a batch of messages',
            type: 'string',
            required: false,
            description: 'the time to wait for a batch of messages'
        }
    ]
};
