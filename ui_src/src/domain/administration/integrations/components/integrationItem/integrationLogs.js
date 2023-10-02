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

import './style.scss';

import React, { useState, useEffect } from 'react';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Copy from '../../../../../components/copy';
import Item from 'antd/lib/list/Item';
import { parsingDate } from '../../../../../services/valueConvertor';
import OverflowTip from '../../../../../components/tooltip/overflowtip';
const IntegrationLogs = ({ integrationName }) => {
    const [logsList, setLogsList] = useState([]);
    const data = [
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        },
        {
            id: 3,
            message: '[INF] Integration slack created successfully\r\n',
            created_at: '2023-10-01T22:49:25.756204+03:00',
            tenant_name: '$memphis'
        }
    ];
    const getIntegrationLogs = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_LOGS}?name=${integrationName}`);
            setLogsList(data);
        } catch (err) {
            return;
        }
    };
    useEffect(() => {
        getIntegrationLogs();
    }, []);

    return (
        <div className="integration-body">
            <div className="integrate-description logs-header">
                <p>Logs Details</p>
                <Copy data={JSON.stringify(logsList)} text="Copy Logs" />
            </div>
            <div className="generic-list-wrapper">
                <div className="list">
                    <div className="coulmns-table">
                        {[
                            {
                                key: '1',
                                title: 'Message',
                                width: '400px'
                            },
                            {
                                key: '2',
                                title: 'Date',
                                width: '200px'
                            }
                        ]?.map((column, index) => {
                            return (
                                <span key={index} style={{ width: column.width }}>
                                    {column.title}
                                </span>
                            );
                        })}
                    </div>
                    <div className="rows-wrapper">
                        {data?.map((row, index) => {
                            return (
                                <div className="pubSub-row" key={index}>
                                    <OverflowTip text={row?.message || row?.tenant_name} width={'400px'}>
                                        {row?.message || row?.tenant_name}
                                    </OverflowTip>

                                    <OverflowTip text={parsingDate(row?.created_at)} width={'200px'}>
                                        {parsingDate(row?.created_at)}
                                    </OverflowTip>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default IntegrationLogs;
