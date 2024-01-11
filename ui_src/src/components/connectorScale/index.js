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

import React, { useEffect, useState } from 'react';
import TitleComponent from 'components/titleComponent';
import InputNumberComponent from 'components/InputNumber';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import { sendTrace } from 'services/genericServices';
import { Form } from 'antd';
import Button from 'components/button';

const ConnectorScale = ({ connector, open, done }) => {
    const [loading, setLoading] = useState(false);
    const [instances, setInstances] = useState(connector?.instances || 1);

    useEffect(() => {
        open && setInstances(connector?.instances || 1);
    }, [open]);

    const onFinish = async () => {
        try {
            setLoading(true);
            try {
                await httpRequest('POST', ApiEndpoints.SCALE_CONNECTOR, {
                    connector_id: connector?.id,
                    instances: instances
                });
                setLoading(false);
            } catch (error) {
                setLoading(false);
            }
            sendTrace('scale', {
                name: connector?.name,
                type: connector?.type,
                connector_type: connector?.connector_type,
                instances: instances
            });
            let connectorData = { ...connector };
            connectorData.instances = instances;
            done(connectorData);
        } catch (err) {
            return;
        }
    };

    return (
        <Form name="form" autoComplete="on" className={'scale-connector'}>
            <Form.Item name="instances" validateTrigger="onChange" initialValue={connector?.instances || 1}>
                <TitleComponent
                    headerTitle={`Scale (${connector?.instances || 1}/15)`}
                    typeTitle="sub-header"
                    headerDescription="Choose the number of the connector instances"
                />
                <InputNumberComponent
                    colorType="black"
                    backgroundColorType="none"
                    fontFamily="Inter"
                    borderColorType="gray"
                    radiusType="semi-round"
                    height="40px"
                    boxShadowsType="none"
                    fontSize="14px"
                    style={{ width: '100%', height: '40px', display: 'flex', alignItems: 'center' }}
                    min={1}
                    max={15}
                    value={instances}
                    onChange={(e) => setInstances(e)}
                    disabled={false}
                />
            </Form.Item>
            <Form.Item>
                <Button
                    placeholder={'Save'}
                    width={'100%'}
                    colorType={'white'}
                    fontSize={'14px'}
                    fontWeight={500}
                    border="none"
                    backgroundColorType={'purple'}
                    onClick={onFinish}
                    radiusType="circle"
                    isLoading={loading}
                    disabled={instances === connector?.instances}
                />
            </Form.Item>
        </Form>
    );
};

export default ConnectorScale;
