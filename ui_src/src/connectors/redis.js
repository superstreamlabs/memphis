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
            name: 'username',
            display: 'Username',
            type: 'string',
            required: false
        },
        {
            name: 'password',
            display: 'Password',
            type: 'string',
            required: false
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
