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

import {
    LOCAL_STORAGE_ALREADY_LOGGED_IN,
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_USER_NAME,
    LOCAL_STORAGE_WELCOME_MESSAGE,
    LOCAL_STORAGE_SKIP_GET_STARTED
} from '../../const/localStorageConsts';
import CreateStationForm from '../../components/createStationForm';

import discordLogo from '../../assets/images/discordLogo.svg';
import githubLogo from '../../assets/images/githubLogo.svg';
import stationImg from '../../assets/images/stationsIconActive.svg';
import docsLogo from '../../assets/images/docsLogo.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import welcome from '../../assets/images/welcome.svg';
import { httpRequest } from '../../services/http';
import { useMediaQuery } from 'react-responsive';
import GenericDetails from './genericDetails';
import FailedStations from './failedStations';
import Loader from '../../components/loader';
import Button from '../../components/button';
import { Context } from '../../hooks/store';
import SysComponents from './sysComponents';
import Modal from '../../components/modal';
import { PRIVACY_URL } from '../../config';
import { Link } from 'react-router-dom';
import GetStarted from './getStarted';
import Throughput from './throughput';
import Resources from './resources';

const Desktop = ({ children }) => {
    const isDesktop = useMediaQuery({ minWidth: 850 });
    return isDesktop ? children : null;
};
const Mobile = ({ children }) => {
    const isMobile = useMediaQuery({ maxWidth: 849 });
    return isMobile ? children : null;
};

const dataSentences = [
    `“Data is the new oil” — Clive Humby`,
    `“With data collection, ‘the sooner the better’ is always the best answer” — Marissa Mayer`,
    `“Data are just summaries of thousands of stories – tell a few of those stories to help make the data meaningful” — Chip and Dan Heath`,
    `“Data really powers everything that we do” — Jeff Weiner`,
    `“Without big data, you are blind and deaf and in the middle of a freeway” — Geoffrey Moore`
];

function OverView() {
    const [state, dispatch] = useContext(Context);
    const [open, modalFlip] = useState(false);
    const createStationRef = useRef(null);
    const [botUrl, SetBotUrl] = useState(require('../../assets/images/bots/avatar1.svg'));
    const [username, SetUsername] = useState('');
    const [isLoading, setisLoading] = useState(true);
    const [creatingProsessd, setCreatingProsessd] = useState(false);

    const [isDataLoaded, setIsDataLoaded] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [showWelcome, setShowWelcome] = useState(false);

    const [dataSentence, setDataSentence] = useState(dataSentences[0]);

    const getRandomInt = (max) => {
        return Math.floor(Math.random() * max);
    };

    const generateSentence = () => {
        setDataSentence(dataSentences[getRandomInt(5)]);
    };

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
        getOverviewData();
        setBotImage(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID) || state?.userData?.avatar_id);
        SetUsername(
            localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== 'undefined' && localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== ''
                ? capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_FULL_NAME))
                : capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_USER_NAME))
        );
        generateSentence();
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
        SetBotUrl(require(`../../assets/images/bots/avatar${botId}.svg`));
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

    return (
        <div className="overview-container">
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}

            {!isLoading && (localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) === 'true' || allStations.length > 0) && (
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
                                    alt="avatar"
                                ></img>
                            </div>
                            <div className="dynamic-sentences">
                                {localStorage.getItem(LOCAL_STORAGE_ALREADY_LOGGED_IN) === 'true' ? <h1>Welcome Back, {username}</h1> : <h1>Welcome, {username}</h1>}
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
                        <div className="left-side">
                            <GenericDetails />
                            <FailedStations />
                            <Throughput />
                        </div>
                        <div className="right-side">
                            <Resources />
                            <SysComponents />
                        </div>
                    </div>
                </div>
            )}
            {!isLoading &&
                (localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) === null || localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) === 'undefined') &&
                allStations.length === 0 && <GetStarted username={username} dataSentence={dataSentence} />}
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <img className="headerImage" src={stationImg} alt="stationImg" />
                        </div>
                        <p>Create new station</p>
                        <label>A station is a distributed unit that stores the produced data.</label>
                    </div>
                }
                height="460px"
                width="540px"
                rBtnText="Add"
                lBtnText="Cancel"
                lBtnClick={() => {
                    modalFlip(false);
                }}
                rBtnClick={() => {
                    createStationRef.current();
                }}
                clickOutside={() => modalFlip(false)}
                open={open}
                isLoading={creatingProsessd}
            >
                <CreateStationForm createStationFormRef={createStationRef} handleClick={(e) => setCreatingProsessd(e)} />
            </Modal>
            <Modal
                header={''}
                height="470px"
                closeAction={() => {
                    setShowWelcome(false);
                    localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, false);
                }}
                clickOutside={() => {
                    setShowWelcome(false);
                    localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, false);
                }}
                open={showWelcome}
                displayButtons={false}
            >
                <div className="sandbox-welcome">
                    <img src={welcome} alt="docs" className="welcome-img"></img>
                    <label className="welcome-header">Welcome aboard</label>
                    <label className="welcome-message">We are super happy to have you with us!</label>
                    <label className="welcome-message">Please remember that this is a sandbox environment</label>
                    <label className="welcome-message">and is under constant modifications.</label>
                    <label className="welcome-message">Downtimes might occur.</label>
                    <div>
                        <Link to={{ pathname: 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/memphis/overview' }} target="_blank">
                            <img src={docsLogo} alt="slack" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://github.com/memphisdev/memphis-broker' }} target="_blank">
                            <img src={githubLogo} alt="github" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://discord.com/invite/WZpysvAeTf' }} target="_blank">
                            <img src={discordLogo} alt="discord" className="sandbox-icon"></img>
                        </Link>
                    </div>
                    <Button
                        width="140px"
                        height="36px"
                        placeholder="Get Started"
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
            </Modal>
        </div>
    );
}

export default OverView;
