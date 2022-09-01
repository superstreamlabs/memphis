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

import React, { useContext, useEffect } from 'react';
import Button from '../../../../components/button';
import SlackIcon from '../../../../assets/images/slackIcon.svg';
import GithubIcon from '../../../../assets/images/githubIcon.svg';
import DiscordIcon from '../../../../assets/images/discordIcon.svg';
import { Link, useHistory } from 'react-router-dom';
import { GetStartedStoreContext } from '..';
import pathDomains from '../../../../router';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';

const Finish = (props) => {
    const { createStationFormRef } = props;

    const history = useHistory();
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);

    useEffect(() => {
        createStationFormRef.current = onNext;
    }, []);

    const onNext = () => {
        doneNextSteps();
        window.location.reload(false);
    };

    const onFinish = (e) => {
        e.preventDefault();
        doneNextSteps();
        history.push(`${pathDomains.factoriesList}/${getStartedState.factoryName}/${getStartedState.stationName}`);
    };

    const doneNextSteps = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.DONE_NEXT_STEPS);
        } catch (error) {}
    };

    return (
        <div className="finish-container">
            <div className="container-icons-finish">
                <Button
                    width="192px"
                    height="42px"
                    placeholder="Go to station"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="16px"
                    fontWeight="bold"
                    border="1px solid #EEEEEE"
                    borderRadius="31px"
                    boxShadowStyle="none"
                    onClick={(e) => {
                        onFinish(e);
                    }}
                />
                <p className="link-finish-header">Link to our channels</p>
                <Link className="icon-image" to={{ pathname: 'https://memphiscommunity.slack.com/archives/C03KRNC6R3Q' }} target="_blank">
                    <img src={SlackIcon} width="25px" height="25px" alt="slack-icon"></img>
                </Link>
                <Link className="icon-image" to={{ pathname: 'https://github.com/memphisdev' }} target="_blank">
                    <img src={GithubIcon} width="25px" height="25px" alt="github-icon"></img>
                </Link>
                <Link className="icon-image" to={{ pathname: 'https://discord.com/invite/WZpysvAeTf' }} target="_blank">
                    <img src={DiscordIcon} width="25px" height="25px" alt="discord_icon"></img>
                </Link>
            </div>
        </div>
    );
};

export default Finish;
