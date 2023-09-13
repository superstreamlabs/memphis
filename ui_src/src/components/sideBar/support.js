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

import { ReactComponent as SupportColorIcon } from '../../assets/images/supportIconColor.svg';
import { ReactComponent as DocumentIcon } from '../../assets/images/documentIcon.svg';
import { ReactComponent as MailsendIcon } from '../../assets/images/mailsendIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import SelectComponent from '../select';
import { DOC_URL } from '../../config';
import EmailLink from '../emailLink';
import Button from '../button';
import Input from '../Input';
import { message } from 'antd';
import { showMessages } from '../../services/genericServices';

const Support = ({ closeModal }) => {
    const [severity, setSeverity] = useState('Critical (Cannot produce or consume data)');
    const [textInfo, setTextInfo] = useState('');
    const [loader, setLoader] = useState(false);

    const clearValues = () => {
        setSeverity('Critical (Cannot produce or consume data)');
        setTextInfo('');
    };

    const sendSupport = async () => {
        setLoader(true);
        const severityValue = severity.split(' ')[0].toLowerCase();
        try {
            await httpRequest('POST', `${ApiEndpoints.SEND_SUPPORT}`, {
                severity: severityValue,
                details: textInfo
            });
            clearValues();
            setTimeout(() => {
                setLoader(false);
                showMessages('success', 'Your ticket has been opened and will be reviewed by our support as soon as possible.');
                closeModal(false);
            }, 1000);
        } catch (error) {
            setLoader(false);
            return;
        }
    };

    return (
        <div className="menu-content">
            <div className="support-container">
                <div className="support-image">
                    <SupportColorIcon />
                </div>
                <p className="popover-header">Need Support?</p>
                <label>We're here to help!</label>
                <p className="support-content-header">If you have any questions or need assistance, please don't hesitate to reach out to our support team.</p>
                <div className="support-span">
                    <div className="support-content">
                        <div className="flex">
                            = <DocumentIcon alt="documentIcon" />
                            <p>Link to Documentation</p>
                        </div>
                        <a href={DOC_URL} target="_blank" rel="noreferrer">
                            Documentation
                        </a>
                    </div>
                    <div className="support-content">
                        <div className="flex">
                            <MailsendIcon alt="mailsendIcon" />
                            <p>Support Email</p>
                        </div>
                        <EmailLink email={'support@memphis.dev'} />
                    </div>
                </div>
                <div>
                    <p className="support-title">Severity</p>
                    <SelectComponent
                        value={severity}
                        colorType="black"
                        backgroundColorType="white"
                        borderColorType="gray-light"
                        radiusType="semi-round"
                        minWidth="12vw"
                        width="350px"
                        height="36px"
                        options={[
                            'Critical (Cannot produce or consume data)',
                            'High (Critical capabilities are not functioning)',
                            'Medium (I can’t get something to work)',
                            'Low / Question'
                        ]}
                        onChange={(value) => {
                            setSeverity(value);
                        }}
                        iconColor="gray"
                        popupClassName="message-option"
                    />
                    <p className="support-title">More information</p>
                    <Input
                        placeholder="Please provide more information"
                        type="textArea"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        numberOfRows={3}
                        fontSize="14px"
                        value={textInfo}
                        onBlur={(e) => setTextInfo(e.target.value)}
                        onChange={(e) => setTextInfo(e.target.value)}
                    />
                </div>
                <div className="close-button">
                    <Button
                        width="170px"
                        height="32px"
                        placeholder="Close"
                        colorType="navy"
                        border="gray"
                        backgroundColorType={'white'}
                        radiusType="circle"
                        fontSize="12px !important"
                        fontFamily="InterSemiBold"
                        onClick={() => {
                            clearValues();
                            closeModal(false);
                        }}
                    />
                    <Button
                        width="170px"
                        height="32px"
                        placeholder="Create a ticket"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        onClick={() => {
                            sendSupport();
                        }}
                        isLoading={loader}
                        disabled={textInfo === ''}
                    />
                </div>
            </div>
        </div>
    );
};

export default Support;
