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

import React, { useState } from 'react';
import Modal from '../../../../../components/modal';
import SlackIntegration from '../slackIntegration';

const IntegrationItem = ({ value }) => {
    const [modalIsOpen, modalFlip] = useState(false);

    const modalContent = () => {
        switch (value.name) {
            case 'Slack':
                return <SlackIntegration close={() => modalFlip(false)} />;
            default:
                break;
        }
    };

    return (
        <>
            <integ-item is="3xd" className="integration-item-container" onClick={() => modalFlip(true)}>
                {value.banner}
                <div className="integration-name">
                    {value.icon}
                    <div className="details">
                        <p>{value.name}</p>
                        <span>by {value.by}</span>
                    </div>
                </div>
                <p className="integration-description">{value.description} </p>
            </integ-item>
            <Modal className="integration-modal" height="95vh" width="650px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                {modalContent()}
            </Modal>
        </>
    );
};

export default IntegrationItem;
