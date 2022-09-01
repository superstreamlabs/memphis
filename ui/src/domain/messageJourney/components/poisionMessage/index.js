// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useState } from 'react';
import Button from '../../../../components/button';
import CustomCollapse from '../../../stationOverview/stationObservabilty/components/customCollapse';
import { Space } from 'antd';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { message as messageAnt } from 'antd';

const PoisionMessage = ({ stationName, messageId, details, message, processing, returnBack }) => {
    const [resendProcess, setResendProcess] = useState(false);
    const [ignoreProcess, setIgnoreProcess] = useState(false);

    const handleIgnore = async () => {
        setIgnoreProcess(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.ACK_POISION_MESSAGE}`, { poison_message_ids: [messageId] });
            setTimeout(() => {
                setIgnoreProcess(false);
                returnBack();
            }, 1500);
        } catch (error) {
            setIgnoreProcess(false);
        }
    };

    const handleResend = async () => {
        setResendProcess(true);
        processing(true);
        try {
            await httpRequest('POST', `${ApiEndpoints.RESEND_POISION_MESSAGE_JOURNEY}`, { poison_message_ids: [messageId] });
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
        <div className="poision-message">
            <header is="x3d">
                <p>
                    {stationName} / #{messageId.substring(0, 5)}
                </p>
                <div className="btn-row">
                    <Button
                        width="75px"
                        height="24px"
                        placeholder="Ignore"
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
                    <CustomCollapse status={false} header="Details" defaultOpen={true} data={details} />
                    <CustomCollapse status={false} header="Payload" defaultOpen={true} data={message} message={true} />
                </Space>
            </div>
        </div>
    );
};
export default PoisionMessage;
