export const memphis = {
    Sink: [
        {
            name: 'name',
            display: 'Connector name',
            type: 'string',
            required: true,
            description: 'Note that the name of the sink connector is also used as the name of the consumer group'
        },
        {
            name: 'route_strategy',
            display: 'Route Strategy',
            type: 'select',
            options: ['Station Name', 'Header'],
            required: true,
            children: true,
            'Station Name': [
                {
                    name: 'station_name',
                    display: 'Station Name',
                    type: 'string',
                    required: true,
                    description: 'The name of the station to route messages to (all messages will be routed to the station)'
                }
            ],
            Header: [
                {
                    name: 'header_name',
                    display: 'Header Name',
                    type: 'string',
                    required: true,
                    description:
                        'The header name in the Memphis message for extracting the station name, with multiple stations indicated by comma-separated values (e.g., station1, station2)'
                }
            ]
        },
        {
            name: 'dest_station_config',
            display: 'Destination Station Config',
            type: 'select',
            options: ['Default Station', 'Custom Station'],
            required: true,
            description: 'if a station dose not exist, it will be created with the chosen configuration - Default / Custom',
            children: true,
            'Default Station': [],
            'Custom Station': [
                {
                    name: 'partition_number',
                    display: 'Partition Number',
                    type: 'string',
                    required: true,
                    description: 'The number of partitions in the station'
                },
                {
                    name: 'retention_policy',
                    display: 'Retention Policy',
                    type: 'select',
                    options: ['Time', 'Size', 'Messages', 'Ack'],
                    required: true,
                    description: 'choose retention policy',
                    children: true,
                    Time: [
                        {
                            name: 'retention_value',
                            display: 'Value',
                            type: 'string',
                            required: true,
                            description: 'retention time value in seconds',
                            placeholder: 0
                        }
                    ],
                    Size: [
                        {
                            name: 'retention_value',
                            display: 'Value',
                            type: 'string',
                            required: true,
                            description: 'retentnion size in bytes',
                            placeholder: 0
                        }
                    ],
                    Messages: [
                        {
                            name: 'retention_value',
                            display: 'Value',
                            type: 'string',
                            required: true,
                            description: 'retention messages number',
                            placeholder: 0
                        }
                    ],
                    Ack: []
                }
            ]
        }
    ]
};
