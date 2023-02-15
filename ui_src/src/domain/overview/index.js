// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useContext, useState, useRef } from 'react';
import { CloudDownloadRounded } from '@material-ui/icons';
import { useMediaQuery } from 'react-responsive';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Link } from 'react-router-dom';

import {
    LOCAL_STORAGE_ALREADY_LOGGED_IN,
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_USER_NAME,
    LOCAL_STORAGE_WELCOME_MESSAGE,
    LOCAL_STORAGE_SKIP_GET_STARTED,
    LOCAL_STORAGE_PROFILE_PIC,
    LOCAL_STORAGE_OVERVIEW_VIEW
} from '../../const/localStorageConsts';
import installationIcon from '../../assets/images/installationIcon.svg';
import stationImg from '../../assets/images/stationsIconActive.svg';
import CreateStationForm from '../../components/createStationForm';
import dashboardView from '../../assets/images/dashboardView.svg';
import { capitalizeFirst } from '../../services/valueConvertor';
import discordLogo from '../../assets/images/discordLogo.svg';
import githubLogo from '../../assets/images/githubLogo.svg';
import SegmentButton from '../../components/segmentButton';
import graphView from '../../assets/images/graphView.svg';
import Installation from '../../components/installation';
import docsLogo from '../../assets/images/docsLogo.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import welcome from '../../assets/images/welcome.svg';
import { httpRequest } from '../../services/http';
import SystemComponents from './systemComponents';
import GenericDetails from './genericDetails';
import FailedStations from './failedStations';
import Loader from '../../components/loader';
import Button from '../../components/button';
import { Context } from '../../hooks/store';
import Modal from '../../components/modal';
import GetStarted from './getStarted';
import Throughput from './throughput';
import GraphView from './graphView';

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

const segOptions = [
    {
        label: 'Dashboard',
        value: 'Dashboard',
        icon: <img src={dashboardView} alt="stationImg" style={{ fill: '#ff0000' }} />
    },
    {
        label: 'Graph view',
        value: 'Graph view',
        icon: <img src={graphView} alt="graphView" style={{ fill: '#ff0000' }} />
    }
];

function OverView() {
    const [state, dispatch] = useContext(Context);
    const [open, modalFlip] = useState(false);
    const createStationRef = useRef(null);
    const [botUrl, SetBotUrl] = useState(require('../../assets/images/bots/avatar1.svg'));
    const [username, SetUsername] = useState('');
    const [isLoading, setisLoading] = useState(true);
    const [creatingProsessd, setCreatingProsessd] = useState(false);
    const [showInstallaion, setShowInstallaion] = useState(false);
    const [showWelcome, setShowWelcome] = useState(false);
    const [overviewView, setOverviewView] = useState(localStorage.getItem(LOCAL_STORAGE_OVERVIEW_VIEW) || 'Dashboard');

    const [dataSentence, setDataSentence] = useState(dataSentences[0]);

    const getRandomInt = (max) => {
        return Math.floor(Math.random() * max);
    };

    const generateSentence = () => {
        setDataSentence(dataSentences[getRandomInt(5)]);
    };

    const arrangeData = (data) => {
        data.stations?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.system_components.sort(function (a, b) {
            let nameA = a.name.toUpperCase();
            let nameB = b.name.toUpperCase();
            if (nameA < nameB) {
                return -1;
            }
            if (nameA > nameB) {
                return 1;
            }
            return 0;
        });
        data.system_components.map((a) => {
            a.ports?.sort(function (a, b) {
                if (a < b) {
                    return -1;
                }
                if (a > b) {
                    return 1;
                }
                return 0;
            });
        });
        dispatch({ type: 'SET_MONITOR_DATA', payload: data });
    };

    const getOverviewData = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_MAIN_OVERVIEW_DATA);
            arrangeData(data);
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
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
        const sc = StringCodec();
        const jc = JSONCodec();
        let sub;
        try {
            (async () => {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.main_overview_data`, sc.encode('SUB'));
                const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                sub = state.socket?.subscribe(`$memphis_ws_pubs.main_overview_data.${brokerName}`);
            })();
        } catch (err) {
            return;
        }
        setisLoading(true);
        setTimeout(async () => {
            if (sub) {
                (async () => {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        arrangeData(data);
                    }
                })();
            }
        }, 1000);
        return () => {
            sub?.unsubscribe();
        };
    }, [state.socket]);

    const setBotImage = (botId) => {
        SetBotUrl(require(`../../assets/images/bots/avatar${botId}.svg`));
    };

    return (
        <div className="overview-container">
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}

            {!isLoading && localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) === 'true' && (
                <div className="overview-wrapper">
                    <div className="header">
                        <div className="header-welcome">
                            <div className="bot-wrapper">
                                <img
                                    className="sandboxUserImg"
                                    src={localStorage.getItem(LOCAL_STORAGE_PROFILE_PIC) || botUrl} // profile_pic is available only in sandbox env
                                    referrerPolicy="no-referrer"
                                    width={localStorage.getItem(LOCAL_STORAGE_PROFILE_PIC) ? 60 : 40}
                                    height={localStorage.getItem(LOCAL_STORAGE_PROFILE_PIC) ? 60 : 40}
                                    alt="avatar"
                                ></img>
                            </div>
                            <div className="dynamic-sentences">
                                {localStorage.getItem(LOCAL_STORAGE_ALREADY_LOGGED_IN) === 'true' ? <h1>Welcome back, {username}</h1> : <h1>Welcome, {username}</h1>}
                                {/* <p className="ok-status">You’re a memphis superhero! All looks good!</p> */}
                            </div>
                        </div>
                        <div className="overview-actions">
                            <SegmentButton
                                value={overviewView}
                                options={segOptions}
                                onChange={(e) => {
                                    setOverviewView(e);
                                    localStorage.setItem(LOCAL_STORAGE_OVERVIEW_VIEW, e);
                                }}
                            />
                            {process.env.REACT_APP_SANDBOX_ENV && (
                                <Button
                                    className="modal-btn"
                                    width="130px"
                                    height="34px"
                                    placeholder={
                                        <div className="title">
                                            <CloudDownloadRounded className="download-icon" />
                                            <p>Install now</p>
                                        </div>
                                    }
                                    colorType="purple"
                                    radiusType="circle"
                                    backgroundColorType="none"
                                    border="purple"
                                    fontSize="12px"
                                    fontWeight="600"
                                    boxShadowStyle="float"
                                    aria-haspopup="true"
                                    onClick={() => setShowInstallaion(true)}
                                />
                            )}
                            <Button
                                className="modal-btn"
                                width="160px"
                                height="34px"
                                placeholder={'Create new station'}
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                boxShadowStyle="float"
                                onClick={() => modalFlip(true)}
                            />
                        </div>
                    </div>
                    {localStorage.getItem(LOCAL_STORAGE_OVERVIEW_VIEW) === 'Dashboard' ? (
                        <div className="overview-components">
                            <div className="left-side">
                                <GenericDetails />
                                <FailedStations createStationTrigger={(e) => modalFlip(e)} />
                                <Throughput />
                            </div>
                            <div className="right-side">
                                <SystemComponents />
                            </div>
                        </div>
                    ) : (
                        <GraphView />
                    )}
                </div>
            )}
            {!isLoading && localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) !== 'true' && <GetStarted username={username} dataSentence={dataSentence} />}
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
                height="65vh"
                width="1020px"
                rBtnText="Create"
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
                        <Link to={{ pathname: 'https://docs.memphis.dev/memphis/getting-started/readme' }} target="_blank">
                            <img src={docsLogo} alt="slack" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://github.com/memphisdev/memphis-broker' }} target="_blank">
                            <img src={githubLogo} alt="github" className="sandbox-icon"></img>
                        </Link>
                        <Link to={{ pathname: 'https://discord.com/invite/3QcAwtrZZR' }} target="_blank">
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
            <Modal
                header={
                    <label className="installation-icon-wrapper">
                        <img src={installationIcon} alt="installationIcon" />
                    </label>
                }
                height="700px"
                clickOutside={() => {
                    setShowInstallaion(false);
                }}
                open={showInstallaion}
                displayButtons={false}
            >
                <Installation closeModal={() => setShowInstallaion(false)} />
            </Modal>
        </div>
    );
}

export default OverView;
