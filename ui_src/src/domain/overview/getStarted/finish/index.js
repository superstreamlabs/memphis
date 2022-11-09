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
import Button from '../../../../components/button';
import Switcher from '../../../../components/switcher';
import docsLogo from '../../../../assets/images/docsLogo.svg';
import GithubLogo from '../../../../assets/images/githubLogo.svg';
import discordLogo from '../../../../assets/images/discordLogo.svg';
import { Link, useHistory } from 'react-router-dom';
import { GetStartedStoreContext } from '..';
import pathDomains from '../../../../router';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_SKIP_GET_STARTED, LOCAL_STORAGE_USER_NAME } from '../../../../const/localStorageConsts';

const Finish = ({ createStationFormRef }) => {
    const history = useHistory();
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [allowAnalytics, setAllowAnalytics] = useState(localStorage.getItem(LOCAL_STORAGE_ALLOW_ANALYTICS) || false);

    useEffect(() => {
        createStationFormRef.current = onNext;
        httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, localStorage.getItem(LOCAL_STORAGE_USER_NAME));
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
        <div className="finish-container" id="e2e-getstarted-step5">
            <div className="btn-container">
                <div className="allow-analytics">
                    <Switcher onChange={() => sendAnalytics(!allowAnalytics)} checked={allowAnalytics} checkedChildren="on" unCheckedChildren="off" />
                    <p>I allow Memphis team to reach out and ask for feedback.</p>
                </div>
                <div id="e2e-getstarted-finish-btn">
                    <Button
                        width="192px"
                        height="42px"
                        placeholder="Go to dashboard"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="16px"
                        fontWeight="bold"
                        colorType="white"
                        borderRadius="31px"
                        boxShadowStyle="none"
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
        </div>
    );
};

export default Finish;
