// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useContext, useState } from 'react';

import integrationRequestIcon from '../../../assets/images/integrationRequestIcon.svg';
import cloudeBadge from '../../../assets/images/cloudeBadge.svg';
import { INTEGRATION_LIST } from '../../../const/integrationList';
import IntegrationItem from './components/integrationItem';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import { Context } from '../../../hooks/store';
import { CloudQueueRounded } from '@material-ui/icons';
import Button from '../../../components/button';
import Modal from '../../../components/modal';
import Input from '../../../components/Input';
import { message } from 'antd';

const Integrations = () => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrationRequest, setIntegrationRequest] = useState('');

    useEffect(() => {
        getallIntegration();
    }, []);

    const getallIntegration = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_INTEGRATION);
            dispatch({ type: 'SET_INTEGRATIONS', payload: data || [] });
        } catch (err) {
            return;
        }
    };
    const handleSendRequest = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.REQUEST_INTEGRATION, { request_content: integrationRequest });
            message.success({
                key: 'memphisSuccessMessage',
                content: 'Thanks for your feedback',
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisSuccessMessage')
            });
            modalFlip(false);
            setIntegrationRequest('');
        } catch (err) {
            return;
        }
    };

    return (
        <div className="alerts-integrations-container">
            <div className="header-preferences">
                <div className="left">
                    <p className="main-header">Integrations</p>
                    <p className="sub-header">Integrations for notifications, monitoring, API calls, and more</p>
                </div>
                <Button
                    width="140px"
                    height="35px"
                    placeholder="Request Integration"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    border="none"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    onClick={() => modalFlip(true)}
                />
            </div>
            <div className="integration-list">
                {INTEGRATION_LIST?.map((integration) =>
                    integration.comingSoon ? (
                        <div key={integration.name} className="cloud-wrapper">
                            <div className="dark-background">
                                <img src={cloudeBadge} />
                                <div className="cloud-icon">
                                    <CloudQueueRounded />
                                </div>
                            </div>
                            <IntegrationItem key={integration.name} value={integration} />
                        </div>
                    ) : (
                        <IntegrationItem key={integration.name} value={integration} />
                    )
                )}
            </div>
            <Modal
                className="request-integration-modal"
                header={<img src={integrationRequestIcon} alt="errorModal" />}
                height="250px"
                width="450px"
                displayButtons={false}
                clickOutside={() => modalFlip(false)}
                open={modalIsOpen}
            >
                <div className="roll-back-modal">
                    <p className="title">Integrations framework</p>
                    <p className="desc">Until our integrations framework will be released, we can build it for you. Which integration is missing?</p>
                    <Input
                        placeholder="App & reason"
                        type="text"
                        fontSize="12px"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        height="40px"
                        onBlur={(e) => setIntegrationRequest(e.target.value)}
                        onChange={(e) => setIntegrationRequest(e.target.value)}
                        value={integrationRequest}
                    />

                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => modalFlip(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Send"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            disabled={integrationRequest === ''}
                            onClick={() => handleSendRequest()}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
};

export default Integrations;
