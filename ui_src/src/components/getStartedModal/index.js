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
import { useState, useEffect } from 'react';
import { BiSolidTimeFive } from 'react-icons/bi';
import VideoPlayer from '../videoPlayer';
import Button from '../button';
import Input from '../Input';
import { WELCOME_VIDEO } from '../../config';
import WelcomeImage from '../../assets/images/welcomeModalImage.webp';
import { BsGithub } from 'react-icons/bs';
import { ReactComponent as CloneModalIcon } from '../../assets/images/cloneModalIcon.svg';
import Modal from '../modal';
import CloneModal from '../cloneModal';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import { sendTrace } from '../../services/genericServices';
import { capitalizeFirst } from '../../services/valueConvertor';
import { LOCAL_STORAGE_SKIP_GET_STARTED, LOCAL_STORAGE_USER_NAME } from '../../const/localStorageConsts';

const useCases = ['Microservices communication', 'Change data Capture', 'Real-time pipeline', 'Stream processing'];
const codeList = [
    {
        title: 'Event-driven Application',
        subtitle: 'To get you up and running with Memphis.dev in no time, we developed Fastmart - The Fastest Food Delivery App Ever Created. Take a look!',
        difficult: 'Easy',
        type: 'MICROSERVICES',
        time: '15-20 Min'
    }
];

const GetStartedModal = ({ open, handleClose }) => {
    const [chosenUseCase, setUseCase] = useState('');
    const [manualUseCase, setManualUseCase] = useState('');
    const [openCloneModal, setOpenCloneModal] = useState(false);

    useEffect(() => {
        sendTrace('user-opened-get-started-modal', {});
    }, []);

    const handleSelectUseCase = (useCase) => {
        setUseCase(useCase);
        setManualUseCase('');
    };

    const handleManualUseCase = (useCase) => {
        setManualUseCase(useCase);
        setUseCase('');
    };

    const skipGetStarted = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, { username: capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_USER_NAME)) });
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, true);
        } catch (error) {
            return;
        }
    };

    const finsihGetStarted = async () => {
        sendTrace('user-chosen-use-case', { use_case: chosenUseCase !== '' ? chosenUseCase : manualUseCase });
        setManualUseCase('');
        if (localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) !== 'true') {
            skipGetStarted();
        }
        handleClose();
    };

    return (
        <Modal
            className="get-started-modal"
            width={'600px'}
            displayButtons={false}
            clickOutside={() => {
                if (localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) !== 'true') {
                    skipGetStarted();
                }
                handleClose();
            }}
            open={open}
        >
            <div>
                <div className="title-wrapper">
                    <p className="title">
                        Welcome to <span> Memphis.dev</span>
                    </p>
                    <p className="sub-title">It’s whole new streaming stack 🚀</p>
                </div>
                <div className="video-wrapper">
                    <VideoPlayer url={WELCOME_VIDEO} bgImg={WelcomeImage} width={'540px'} height={'250px'} />
                </div>
                <div className="modal-titles">
                    <label className="title">1. Tell us what brings you to Memphis.dev. We will help accordingly</label>
                </div>
                <use-cases is="x3s">
                    <div className="use-cases">
                        {useCases?.map((useCase, index) => {
                            return (
                                <span className={chosenUseCase === useCase && 'selected'} key={index} onClick={() => handleSelectUseCase(useCase)}>
                                    {useCase}
                                </span>
                            );
                        })}
                    </div>
                    <Input
                        value={manualUseCase}
                        placeholder="Tell us about your use case"
                        type="text"
                        radiusType="semi-round"
                        borderColorType="gray-light"
                        boxShadowsType="none"
                        colorType="black"
                        backgroundColorType="none"
                        width="100%"
                        minWidth="360px"
                        height="40px"
                        iconComponent=""
                        onChange={(e) => handleManualUseCase(e.target.value)}
                    />
                </use-cases>
                <div className="modal-titles">
                    <label className="title">2. Start your journey with our onboarding application</label>
                </div>
                {/* <Divider plain>
                    Or start with an <label className="example-app">example application</label>
                </Divider> */}
                {codeList?.map((code, index) => {
                    return (
                        <tutorial
                            is="x3s"
                            key={index}
                            onClick={() => {
                                setOpenCloneModal(true);
                                sendTrace('user-click-example-app', { app: code.title });
                            }}
                        >
                            <div className="left-purple"></div>
                            <data is="x3s">
                                <header is="x3s">
                                    <label>TUTORIAL</label>
                                    <BsGithub
                                        alt="github"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            window.open('https://github.com/memphisdev/onboarding-app.git', '_blank');
                                        }}
                                    />
                                </header>
                                <body is="x3s">
                                    <label className="title">{code.title}</label>
                                    <label className="subtitle">{code.subtitle}</label>
                                </body>
                                <footer is="x3s">
                                    <info is="x3s">
                                        <label className="difficult">{code.difficult}</label>
                                        <label className="type">{code.type}</label>
                                    </info>
                                    <time is="x3s">
                                        <BiSolidTimeFive />
                                        <label>{code.time}</label>
                                    </time>
                                </footer>
                            </data>
                        </tutorial>
                    );
                })}

                <div className="footer">
                    <Button
                        width={'200px'}
                        height={'34px'}
                        placeholder={`Let's go!`}
                        colorType={'white'}
                        backgroundColorType={'purple'}
                        radiusType={'circle'}
                        onClick={() => finsihGetStarted()}
                        fontWeight={600}
                        fontSize={'12px'}
                    />
                </div>
            </div>
            <Modal
                header={<CloneModalIcon alt="cloneModalIcon" />}
                width="540px"
                displayButtons={false}
                clickOutside={() => {
                    setManualUseCase('');
                    setOpenCloneModal(false);
                }}
                open={openCloneModal}
            >
                <CloneModal />
            </Modal>
        </Modal>
    );
};

export default GetStartedModal;
