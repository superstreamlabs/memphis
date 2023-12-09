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

import React, { useState } from 'react';
import Button from '../../../../components/button';
import CustomCollapse from '../../../stationOverview/stationObservabilty/components/customCollapse';
import { Space } from 'antd';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { message as messageAnt } from 'antd';

const PoisonMessage = ({ stationName, messageId, details, message, headers, processing, returnBack, schemaType }) => {
    const [resendProcess, setResendProcess] = useState(false);
    const [ignoreProcess, setIgnoreProcess] = useState(false);

    const handleIgnore = async () => {
        setIgnoreProcess(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.DROP_DLS_MESSAGE}`, { dls_type: 'poison', dls_message_ids: [Number(messageId)], station_name: stationName });
            setTimeout(() => {
                setIgnoreProcess(false);
                returnBack();
                messageAnt.success({
                    key: 'memphisSuccessMessage',
                    content: 'The message was dropped successfully',
                    duration: 5,
                    style: { cursor: 'pointer' },
                    onClick: () => message.destroy('memphisSuccessMessage')
                });
            }, 1500);
        } catch (error) {
            setIgnoreProcess(false);
        }
    };

    const handleResend = async () => {
        setResendProcess(true);
        processing(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.RESEND_POISON_MESSAGE_JOURNEY}`, { poison_message_ids: [Number(messageId)], station_name: stationName });
            setTimeout(() => {
                setResendProcess(false);
                processing(false);
                messageAnt.success({
                    key: 'memphisSuccessMessage',
                    content: 'The message was sent successfully',
                    duration: 5,
                    style: { cursor: 'pointer' },
                    onClick: () => message.destroy('memphisSuccessMessage')
                });
            }, 1500);
        } catch (error) {
            setResendProcess(false);
            processing(false);
        }
    };

    return (
        <div className="poison-message">
            <header is="x3d">
                <p>Unacknowledged message details</p>
                <div className="btn-row">
                    <Button
                        width="75px"
                        height="24px"
                        placeholder="Drop"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        isLoading={ignoreProcess}
                        onClick={() => handleIgnore()}
                    />
                    <Button
                        width="90px"
                        height="24px"
                        placeholder="Resend"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        isLoading={resendProcess}
                        onClick={() => handleResend()}
                    />
                </div>
            </header>
            <div className="content-wrapper">
                <Space direction="vertical">
                    <CustomCollapse status={false} header="Metadata" defaultOpen={true} data={details} />
                    <CustomCollapse status={false} header="Headers" defaultOpen={true} data={headers} message={true} />
                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={message} message={true} schemaType={schemaType} />
                </Space>
            </div>
        </div>
    );
};
export default PoisonMessage;
