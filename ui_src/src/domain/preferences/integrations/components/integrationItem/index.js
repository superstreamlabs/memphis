// Credit for The NATS.IO Authors
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

import React, { useState, useEffect, useContext, useRef } from 'react';

import integrateIcon from '../../../../../assets/images/integrateIcon.svg';
import { capitalizeFirst } from '../../../../../services/valueConvertor';
import { Context } from '../../../../../hooks/store';
import Modal from '../../../../../components/modal';
import SlackIntegration from '../slackIntegration';

const IntegrationItem = ({ value }) => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrateValue, setIntegrateValue] = useState({});

    const ref = useRef();
    ref.current = integrateValue;

    useEffect(() => {
        if (state.integrationsList?.length > 0) {
            checkIfUsed();
        }
    }, [state?.integrationsList]);

    const checkIfUsed = () => {
        let index = state.integrationsList?.findIndex((integration) => capitalizeFirst(integration.name) === value.name);
        setIntegrateValue(state.integrationsList[index]);
    };

    const modalContent = () => {
        switch (value.name) {
            case 'Slack':
                return (
                    <SlackIntegration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                    />
                );
            default:
                break;
        }
    };

    return (
        <>
            <integ-item is="3xd" className="integration-item-container" onClick={() => modalFlip(true)}>
                {value.banner}
                {integrateValue && Object.keys(integrateValue)?.length !== 0 && (
                    <div className="integrate-icon">
                        <img src={integrateIcon} />
                    </div>
                )}
                <div className="integration-name">
                    {value.icon}
                    <div className="details">
                        <p>{value.name}</p>
                        <span>by {value.by}</span>
                    </div>
                </div>
                <p className="integration-description">{value.description} </p>
            </integ-item>
            <Modal className="integration-modal" height="95vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                {modalContent()}
            </Modal>
        </>
    );
};

export default IntegrationItem;
