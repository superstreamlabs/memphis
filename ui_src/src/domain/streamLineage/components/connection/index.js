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

import React from 'react';
import { ReactComponent as ProducerIcon } from 'assets/images/producerIcon.svg';
import { ReactComponent as KafkaIcon } from 'connectors/assets/kafkaIcon.svg';
import { ReactComponent as RedisIcon } from 'connectors/assets/redisIcon.svg';
import { ReactComponent as MemphisIcon } from 'connectors/assets/memphisIcon.svg';

const Connection = ({ id, producer, consumer }) => {

    function getIconByLang() {
        const source = producer || consumer;
        const connector_type = source?.connector_details?.connector_type;

        if (source?.type !== "connector" || !connector_type) return

        const sourceIcons = {
            kafka: <KafkaIcon/>,
            redis: <RedisIcon/>,
            memphis: <MemphisIcon/>,
        };

        const iconComponent = connector_type ? sourceIcons[connector_type] : <ProducerIcon />;

        return <div style={{ fontSize: '17px', display: 'flex', alignItems: 'center' }}>{iconComponent}</div>;
    }

    return (
        <div className="connection-wrapper">
            {producer && (
                <div key={id} className="rectangle producer">
                    <div style={{marginRight: '5px'}}>
                        {getIconByLang()}
                    </div>
                    <p>{producer.name}</p>
                    <div className="count">{producer.count}</div>
                </div>
            )}
            {consumer && (
                <div key={id} className="rectangle consumer">
                    <div style={{marginRight: '5px'}}>
                        {getIconByLang()}
                    </div>
                    <p>{consumer.name}</p>
                    <div className="count">{consumer.count}</div>
                </div>
            )}
        </div>
    );
};

export default Connection;
