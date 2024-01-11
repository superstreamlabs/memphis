export const redis = {
    Sink: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true,
            description: 'Note that the name of the sink connector is also used as the name of the consumer group'
        },
        {
            name: 'redis_type',
            display: 'Type',
            type: 'select',
            options: ['KV'],
            required: true
            // placeholder: 'type'
        },
        {
            name: 'addr',
            display: 'Addr',
            type: 'string',
            placeholder: 'host:port',
            required: true,
            description: '',
            children: false
        },
        {
            name: 'authentication',
            display: 'Authentication method',
            type: 'select',
            options: ['No authentication', 'Username and Password', 'Password Only'],
            required: true,
            children: true,
            'Username and Password': [
                {
                    name: 'username',
                    display: 'Username',
                    type: 'string',
                    required: true
                },
                {
                    name: 'password',
                    display: 'Password',
                    type: 'string',
                    required: true
                }
            ],
            'Password Only': [
                {
                    name: 'password',
                    display: 'Password',
                    type: 'string',
                    required: true
                }
            ],
            'No authentication': []
        },
        {
            name: 'db',
            display: 'Database',
            type: 'string',
            required: true,
            placeholder: 0,
            description: 'The Redis database number to use'
        },
        {
            name: 'key_header',
            display: 'Key header',
            description: 'The header name in the Memphis message from which the Redis key is derived',
            type: 'string',
            required: true
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
            display: 'Batch Message Timeout Duration (Seconds)',
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
