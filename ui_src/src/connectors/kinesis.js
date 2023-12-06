export const kinesis = {
    source: [
        {
            name: 'name',
            display: 'Name',
            type: 'string',
            required: true
        },
        {
            name: 'access_key',
            display: 'access_key',
            type: 'string',
            required: false
        },
        {
            name: 'secret_key',
            display: 'secret_key',
            type: 'string',
            required: false
        },
        {
            name: 'role',
            display: 'role',
            type: 'string',
            required: true
        },
        {
            name: 'kinesis_stream_name',
            display: 'kinesis_stream_name',
            type: 'string',
            required: true
        },
        {
            name: 'shard_iterator_type',
            display: 'shard_iterator_type',
            type: 'select',
            options: ['LATEST', 'TRIM_HORIZON', 'AT_SEQUENCE_NUMBER', 'AFTER_SEQUENCE_NUMBER', 'AT_TIMESTAMP'],
            required: true
        },
        {
            name: 'shard_id',
            display: 'shard_id',
            type: 'string',
            required: true
        },
        {
            name: 'region',
            display: 'region',
            type: 'string',
            required: true
        }
    ],
    sink: [
        {
            name: 'name',
            display: 'Name',
            type: 'string',
            required: true
        },
        {
            name: 'access_key',
            display: 'access_key',
            type: 'string',
            required: false
        },
        {
            name: 'secret_key',
            display: 'secret_key',
            type: 'string',
            required: false
        },
        {
            name: 'role',
            display: 'role',
            type: 'string',
            required: true
        },
        {
            name: 'region',
            display: 'region',
            type: 'string',
            required: true
        },
        {
            name: 'kinesis_stream_name',
            display: 'kinesis_stream_name',
            type: 'string',
            required: true
        },
        {
            name: 'partition_key',
            display: 'partition_key',
            type: 'string',
            required: true
        }
    ]
};
