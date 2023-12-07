export const kafka = {
    source: [
        {
            name: 'name',
            display: 'name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap.servers',
            display: 'bootstrap.servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'localhost:9092',
            description: 'list of brokers'
        },
        {
            name: 'security.protocol',
            display: 'security.protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl.mechanism',
                    display: 'ssl.mechanism',
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'ssl.certificate.pem',
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.key.password',
                    display: 'ssl.key.password',
                    type: 'string',
                    required: true
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl.mechanism',
                    display: 'sasl.mechanism',
                    type: 'select',
                    options: ['PLAIN', 'SCRAM-SHA-256'],
                    required: true
                },
                {
                    name: 'sasl.username',
                    display: 'sasl.username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl.password',
                    display: 'sasl.password',
                    type: 'string',
                    required: true
                }
            ],
            'No authentication': []
        },
        {
            name: 'group.id',
            display: 'group.id',
            type: 'string',
            required: true,
            description: 'consumer group id'
        },
        {
            name: 'offset',
            display: 'offset',
            type: 'string',
            required: false,
            description: 'earliest / end / specific offset (int)'
        },
        {
            name: 'topic',
            display: 'topic',
            type: 'string',
            required: true,
            description: 'topic name'
        },
        {
            name: 'partition',
            display: 'partition',
            type: 'string',
            required: false,
            description: 'partition number'
        }
    ],
    sink: [
        {
            name: 'name',
            display: 'name',
            type: 'string',
            required: true
        },
        {
            name: 'bootstrap.servers',
            display: 'bootstrap.servers',
            type: 'multi',
            options: [],
            required: true,
            placeholder: 'localhost:9092',
            description: 'list of brokers'
        },
        {
            name: 'security.protocol',
            display: 'security.protocol',
            type: 'select',
            options: ['SSL', 'SASL_SSL', 'No authentication'],
            required: true,
            description: 'SSL / SASL_SSL / No authentication',
            children: true,
            SSL: [
                {
                    name: 'ssl.mechanism',
                    display: 'ssl.mechanism',
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.certificate.pem',
                    display: 'ssl.certificate.pem',
                    type: 'string',
                    required: true
                },
                {
                    name: 'ssl.key.password',
                    display: 'ssl.key.password',
                    type: 'string',
                    required: true
                }
            ],
            SASL_SSL: [
                {
                    name: 'sasl.mechanism',
                    display: 'sasl.mechanism',
                    type: 'select',
                    options: ['PLAIN', 'SCRAM-SHA-256'],
                    required: true
                },
                {
                    name: 'sasl.username',
                    display: 'sasl.username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'sasl.password',
                    display: 'sasl.password',
                    type: 'string',
                    required: true
                }
            ],
            'No authentication': []
        },
        {
            name: 'offset',
            display: 'offset',
            type: 'string',
            required: false,
            description: 'earliest / end / specific offset (int)'
        },
        {
            name: 'topic',
            display: 'topic',
            type: 'string',
            required: true,
            description: 'topic name'
        },
        {
            name: 'partition',
            display: 'partition',
            type: 'string',
            required: false,
            description: 'partition number'
        }
    ]
};
