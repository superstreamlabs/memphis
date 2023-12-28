export const redis = {
    Sink: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true,
            description: 'Note that the sink connector name is also consumer group name'
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
            display: 'Authentication',
            type: 'select',
            options: ['No authentication', 'Username and Password', 'Password Only'],
            required: true,
            description: 'No Authentication, Username and Password, Password Only',
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
            display: 'DB',
            type: 'string',
            required: true,
            placeholder: 0,
            description: 'The Redis database number to use'
        },
        {
            name: 'key_header',
            display: 'Key header',
            description: 'The name of the header in Memphis message, to take the Redis key from',
            type: 'string',
            required: true
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
            placeholder: 2,
            type: 'string',
            required: false,
            description: 'The duration which a batch of messages is awaited till processing'
        }
    ]
};
