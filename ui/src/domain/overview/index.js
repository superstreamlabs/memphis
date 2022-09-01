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

import React, { useEffect, useContext, useState, useRef } from 'react';

import { useMediaQuery } from 'react-responsive';
import FailedStations from './failedStations';
import GenericDetails from './genericDetails';
import SysComponents from './sysComponents';
import { Context } from '../../hooks/store';
import Throughput from './throughput';
import Resources from './resources';

import docsIcon from '../../assets/images/docsIcon.png';
import discordIcon from '../../assets/images/discordColor.png';
import slackIcon from '../../assets/images/slackColor.png';

import Button from '../../components/button';
import CreateStationDetails from '../../components/createStationDetails';
import Modal from '../../components/modal';
import { LOCAL_STORAGE_ALREADY_LOGGED_IN, LOCAL_STORAGE_AVATAR_ID, LOCAL_STORAGE_USER_NAME, LOCAL_STORAGE_WELCOME_MESSAGE } from '../../const/localStorageConsts';
import { PRIVACY_URL } from '../../config';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import Loader from '../../components/loader';
import GetStarted from './getStarted';
import { Link } from 'react-router-dom';

const Desktop = ({ children }) => {
    const isDesktop = useMediaQuery({ minWidth: 850 });
    return isDesktop ? children : null;
};
const Mobile = ({ children }) => {
    const isMobile = useMediaQuery({ maxWidth: 849 });
    return isMobile ? children : null;
};

function OverView() {
    const [state, dispatch] = useContext(Context);
    const [open, modalFlip] = useState(false);
    const createStationRef = useRef(null);
    const [botUrl, SetBotUrl] = useState(require('../../assets/images/bots/1.svg'));
    const [username, SetUsername] = useState('');
    const [isLoading, setisLoading] = useState(false);
    const [isDataLoaded, setIsDataLoaded] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [showWelcome, setShowWelcome] = useState(true);
    const [welcomeMessage, setWelcomeMessage] = useState('');

    const getOverviewData = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_MAIN_OVERVIEW_DATA);
            data.stations?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
            dispatch({ type: 'SET_MONITOR_DATA', payload: data });
            setisLoading(false);
            setIsDataLoaded(true);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        getAllStations();
        dispatch({ type: 'SET_ROUTE', payload: 'overview' });
        setShowWelcome(process.env.REACT_APP_SANDBOX_ENV && localStorage.getItem(LOCAL_STORAGE_WELCOME_MESSAGE) === 'true');
        setWelcomeMessage(process.env.REACT_APP_SANDBOX_ENV ? 'Hey ' + capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_USER_NAME)) + ',' : '');
        getOverviewData();
        setBotImage(state?.userData?.avatar_id || localStorage.getItem(LOCAL_STORAGE_AVATAR_ID));
        SetUsername(localStorage.getItem(LOCAL_STORAGE_USER_NAME));
    }, []);

    useEffect(() => {
        state.socket?.on('main_overview_data', (data) => {
            data.stations?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
            dispatch({ type: 'SET_MONITOR_DATA', payload: data });
        });

        setTimeout(() => {
            state.socket?.emit('register_main_overview_data');
            setisLoading(false);
        }, 1000);
        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const setBotImage = (botId) => {
        SetBotUrl(require(`../../assets/images/bots/${botId}.svg`));
    };

    const capitalizeFirst = (str) => {
        return str.charAt(0).toUpperCase() + str.slice(1);
    };

    const getAllStations = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_ALL_STATIONS}`);
            setAllStations(res);
        } catch (err) {
            return;
        }
    };

    const stationsOfUser = allStations.filter((station) => station.created_by_user === username);
    return (
        <div className="overview-container">
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!isLoading && (
                <div className="overview-wrapper">
                    <div className="header">
                        <div className="header-welcome">
                            <div className="bot-wrapper">
                                <img
                                    className="sandboxUserImg"
                                    src={localStorage.getItem('profile_pic') || botUrl} // profile_pic is available only in sandbox env
                                    referrerPolicy="no-referrer"
                                    width={localStorage.getItem('profile_pic') ? 60 : 40}
                                    height={localStorage.getItem('profile_pic') ? 60 : 40}
                                    alt="bot"
                                ></img>
                            </div>
                            <div className="dynamic-sentences">
                                {localStorage.getItem(LOCAL_STORAGE_ALREADY_LOGGED_IN) === 'true' ? (
                                    <h1>Welcome Back, {username}</h1>
                                ) : (
                                    <h1>Welcome Aboard, {username}</h1>
                                )}
                                {/* <p className="ok-status">You’re a memphis superhero! All looks good!</p> */}
                            </div>
                        </div>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="36px"
                            placeholder={'Create new station'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => modalFlip(true)}
                        />
                    </div>
                    <div className="overview-components">
                        {stationsOfUser.length === 0 ? (
                            <div className="left-side">
                                <GetStarted />
                            </div>
                        ) : (
                            <div className="left-side">
                                <GenericDetails />
                                <FailedStations />
                                <Throughput />
                            </div>
                        )}
                        <div className="right-side">
                            <Resources />
                            <SysComponents />
                        </div>
                    </div>
                </div>
            )}
            <Modal
                header="Your station details"
                minHeight="610px"
                minWidth="500px"
                rBtnText="Add"
                lBtnText="Cancel"
                closeAction={() => modalFlip(false)}
                lBtnClick={() => {
                    modalFlip(false);
                }}
                rBtnClick={() => {
                    createStationRef.current();
                }}
                clickOutside={() => modalFlip(false)}
                open={open}
            >
                <CreateStationDetails chooseFactoryField={true} createStationRef={createStationRef} />
            </Modal>
            <Modal
                header={''}
                height="300px"
                minWidth="645px"
                hr={false}
                closeAction={() => {
                    setShowWelcome(false);
                    localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, false);
                }}
                clickOutside={() => {
                    setShowWelcome(false);
                    localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, false);
                }}
                open={showWelcome}
            >
                <div className="sandbox-welcome">
                    <label className="welcome-header">{welcomeMessage}</label>
                    <br />
                    <label>
                        We are super happy to have you with us!
                        <br />
                        Please remember that this is a sandbox environment and is under constant modifications.
                        <br />
                        Downtimes might occur.
                    </label>
                    <br />
                    <div className="sandbox-welcome-links">
                        <Link to={{ pathname: 'https://docs.memphis.dev' }} target="_blank">
                            <img src={docsIcon} alt="docs" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://join.slack.com/t/memphiscommunity/shared_invite/zt-1bdp9ydfk-QpwYIOTz4nkvTGtEL6kJYQ' }} target="_blank">
                            <img src={slackIcon} alt="slack" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://discord.com/invite/WZpysvAeTf' }} target="_blank">
                            {'  '}
                            <img src={discordIcon} alt="discord" className="sandbox-icon"></img>
                        </Link>
                    </div>
                    <div className="welcome-modal-btn">
                        <Button
                            width="130px"
                            height="36px"
                            placeholder={'Get started'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => {
                                setShowWelcome(false);
                                localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, false);
                            }}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default OverView;
