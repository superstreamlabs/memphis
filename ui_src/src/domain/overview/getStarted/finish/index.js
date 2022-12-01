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

import React, { useContext, useEffect, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';

import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_SKIP_GET_STARTED, LOCAL_STORAGE_USER_NAME } from '../../../../const/localStorageConsts';
import slackColors from '../../../../assets/images/slackColors.svg';
import discordLogo from '../../../../assets/images/discordLogo.svg';
import GithubLogo from '../../../../assets/images/githubLogo.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import docsLogo from '../../../../assets/images/docsLogo.svg';
import { httpRequest } from '../../../../services/http';
import Switcher from '../../../../components/switcher';
import Button from '../../../../components/button';
import { GetStartedStoreContext } from '..';
import pathDomains from '../../../../router';
import Modal from '../../../../components/modal';
import SlackIntegration from '../../../preferences/integrations/components/slackIntegration';

const Finish = ({ createStationFormRef }) => {
    const history = useHistory();
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [allowAnalytics, setAllowAnalytics] = useState(localStorage.getItem(LOCAL_STORAGE_ALLOW_ANALYTICS) || false);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrateValue, setIntegrateValue] = useState({});

    useEffect(() => {
        createStationFormRef.current = onNext;
        httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, localStorage.getItem(LOCAL_STORAGE_USER_NAME));
        getIntegration();
    }, []);

    const onNext = () => {
        doneNextSteps();
        window.location.reload(false);
    };

    const onFinish = (e) => {
        e.preventDefault();
        getStartedDispatch({ type: 'INITIAL_STATE', payload: {} });
        doneNextSteps();
        history.push(`${pathDomains.stations}/${getStartedState.stationName}`);
        localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, true);
    };

    const doneNextSteps = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.DONE_NEXT_STEPS);
        } catch (error) {}
    };

    const getIntegration = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=slack`);
            setIntegrateValue(data);
        } catch (error) {}
    };

    const sendAnalytics = async (analyticsFlag) => {
        try {
            await httpRequest('PUT', `${ApiEndpoints.EDIT_ANALYTICS}`, { send_analytics: analyticsFlag });
            setAllowAnalytics(analyticsFlag);
            localStorage.setItem(LOCAL_STORAGE_ALLOW_ANALYTICS, analyticsFlag);
        } catch (err) {
            return;
        }
    };

    return (
        <div className="finish-container">
            <div className="btn-container">
                <div className="allow-analytics">
                    <Switcher onChange={() => sendAnalytics(!allowAnalytics)} checked={allowAnalytics} checkedChildren="on" unCheckedChildren="off" />
                    <p>I allow Memphis team to reach out and ask for feedback.</p>
                </div>
                <div className="buttons-wrapper">
                    <Button
                        height="42px"
                        placeholder={
                            <div className="slack-button">
                                <img src={slackColors} />
                                <p>Integrate Slack</p>
                            </div>
                        }
                        radiusType="circle"
                        backgroundColorType="white"
                        colorType="black"
                        border={'gray-light'}
                        borderRadius="31px"
                        boxShadowStyle="none"
                        marginTop="20px"
                        onClick={() => {
                            modalFlip(true);
                        }}
                    />
                    <Button
                        height="42px"
                        placeholder="Go to station overview"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="16px"
                        fontWeight="bold"
                        colorType="white"
                        borderRadius="31px"
                        boxShadowStyle="none"
                        marginTop="20px"
                        onClick={(e) => {
                            onFinish(e);
                        }}
                    />
                </div>
            </div>
            <div className="container-icons-finish">
                <p className="link-finish-header">Link to our channels</p>
                <Link
                    className="icon-image"
                    to={{ pathname: 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/getting-started/1-installation' }}
                    target="_blank"
                >
                    <img src={docsLogo} width="25px" height="25px" alt="slack-icon"></img>
                </Link>
                <Link className="icon-image" to={{ pathname: 'https://github.com/memphisdev' }} target="_blank">
                    <img src={GithubLogo} width="25px" height="25px" alt="github-icon"></img>
                </Link>
                <Link className="icon-image" to={{ pathname: 'https://discord.com/invite/WZpysvAeTf' }} target="_blank">
                    <img src={discordLogo} width="25px" height="25px" alt="discord_icon"></img>
                </Link>
            </div>
            <Modal className="integration-modal" height="95vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                <SlackIntegration
                    close={(data) => {
                        modalFlip(false);
                        setIntegrateValue(data);
                    }}
                    value={integrateValue}
                />
            </Modal>
        </div>
    );
};

export default Finish;
