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
import { useState } from 'react';
import { BiSolidTimeFive } from 'react-icons/bi';
import VideoPlayer from '../videoPlayer';
import Button from '../button';
import Input from '../Input';
import { CONNECT_APP_VIDEO } from '../../config';
import ConnectBG from '../../assets/images/connectBG.webp';
import { ReactComponent as CloneModalIcon } from '../../assets/images/cloneModalIcon.svg';
import Modal from '../modal';
import CloneModal from '../cloneModal';
import { Divider } from 'antd';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import { capitalizeFirst } from '../../services/valueConvertor';
import { LOCAL_STORAGE_SKIP_GET_STARTED, LOCAL_STORAGE_USER_NAME } from '../../const/localStorageConsts';

const useCases = ['Microservices communication', 'Change data Capture', 'Real-time pipeline', 'Building a data lake'];
const codeList = [
    {
        title: 'Real-time pipeline',
        subtitle:
            'This guide will teach you how to use Bytewax to aggregate on a custom session window on streaming data using reduce and then calculate metrics downstream.',
        difficult: 'Easy',
        type: 'MICROSERVICES',
        time: '15-20 Min'
    }
];

const GetStartedModal = ({ open, handleClose }) => {
    const [chosenUseCase, setUseCase] = useState('');
    const [manualUseCase, setManualUseCase] = useState('');
    const [openCloneModal, setOpenCloneModal] = useState(false);

    const handleSelectUseCase = (useCase) => {
        setUseCase(useCase);
        setManualUseCase('');
    };

    const handleManualUseCase = (useCase) => {
        setManualUseCase(useCase);
        setUseCase('');
    };

    const userChosenUseCase = async () => {
        const bodyRequest = {
            trace_name: 'user-chosen-use-case',
            trace_params: {
                use_case: chosenUseCase !== '' ? chosenUseCase : manualUseCase
            }
        };
        try {
            await httpRequest('POST', ApiEndpoints.SEND_TRACE, bodyRequest);
        } catch (error) {
            return;
        }
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
        userChosenUseCase();
        if (localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) !== 'true') {
            skipGetStarted();
        }
    };

    return (
        <Modal
            className="get-started-modal"
            width={'600px'}
            height={'95vh'}
            displayButtons={false}
            clickOutside={() => {
                skipGetStarted();
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
                    <VideoPlayer url={CONNECT_APP_VIDEO} bgImg={ConnectBG} width={'540px'} height={'250px'} />
                </div>
                <use-cases is="x3s">
                    <div className="header">
                        <label className="title">Tell us what brings you to Memphis.dev today</label>
                    </div>
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
                <Divider plain>
                    Or start with an <label className="example-app">example application</label>
                </Divider>
                {codeList?.map((code, index) => {
                    return (
                        <tutorial is="x3s" key={index} onClick={() => setOpenCloneModal(true)}>
                            <div className="left-purple"></div>
                            <data is="x3s">
                                <header is="x3s">TUTORIAL</header>
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
                clickOutside={() => setOpenCloneModal(false)}
                open={openCloneModal}
            >
                <CloneModal />
            </Modal>
        </Modal>
    );
};

export default GetStartedModal;
