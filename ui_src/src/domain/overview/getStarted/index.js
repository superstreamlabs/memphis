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

import React, { createContext, useEffect, useReducer, useRef, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Divider } from 'antd';
import { LOCAL_STORAGE_SKIP_GET_STARTED, LOCAL_STORAGE_USER_NAME, LOCAL_STORAGE_USER_PASS_BASED_AUTH } from '../../../const/localStorageConsts';
import GetStartedItem from '../../../components/getStartedItem';
import GetStartedIcon from '../../../assets/images/getStartedIcon.svg';
import AppUserIcon from '../../../assets/images/usersIconActive.svg';
import ProduceDataImg from '../../../assets/images/emptyStation.svg';
import ConsumeDataImg from '../../../assets/images/fullStation.svg';
import finishStep from '../../../assets/images/readyToRoll.svg';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import ProduceConsumeData from './produceConsumeData';
import { httpRequest } from '../../../services/http';
import Button from '../../../components/button';
import SkipGetStrtedModal from '../../../components/skipGetStartedModal';
import CreateAppUser from './createAppUser';
import CreateStation from './createStation';
import Reducer from './hooks/reducer';
import SideStep from './sideStep';
import Finish from './finish';
import { capitalizeFirst, isCloud } from '../../../services/valueConvertor';

const steps = [{ stepName: 'Create Station' }, { stepName: 'Create App user' }, { stepName: 'Produce data' }, { stepName: 'Consume data' }, { stepName: 'Finish' }];

const finishStyle = {
    container: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center'
    },
    header: {
        fontFamily: 'Inter',
        fontStyle: 'normal',
        fontWeight: 600,
        fontSize: '24px',
        lineHeight: '29px',
        color: '#1D1D1D'
    },
    description: {
        fontFamily: 'Inter',
        fontStyle: 'normal',
        fontWeight: 400,
        fontSize: '14px',
        lineHeight: '120%',
        textAlign: 'center',
        color: '#B4B4B4'
    },
    image: {
        width: '150px',
        height: '150px'
    }
};

const initialState = {
    currentStep: 1,
    completedSteps: 0,
    formFieldsCreateStation: {
        name: '',
        retention_type: 'message_age_sec',
        retention_value: 3600,
        storage_type: 'file',
        replicas: 'No HA (1)',
        partitions_number: 1,
        days: 1,
        hours: 0,
        minutes: 0,
        seconds: 0,
        retentionSizeValue: '1000',
        retentionMessagesValue: '10',
        tiered_storage_enabled: false
    },
    username: '',
    password: '',
    nextDisable: false,
    isLoading: false,
    isHiddenButton: false,
    actualPods: null
};

const GetStarted = ({ username, dataSentence, skip }) => {
    const [getStartedState, getStartedDispatch] = useReducer(Reducer, initialState);
    const [open, modalFlip] = useState(false);
    const history = useHistory();
    const createStationFormRef = useRef(null);
    const getStartedStateRef = useRef(null);
    const [targetLocation, setTargetLocation] = useState(null);
    const [displayGetStarted, setDisplayGetStarted] = useState(true);

    useEffect(() => {
        if (!displayGetStarted && targetLocation !== null) {
            history.push(targetLocation);
            setTargetLocation(null);
        }
    }, [displayGetStarted, targetLocation, history]);

    useEffect(() => {
        getStartedStateRef.current = getStartedState;
    }, [getStartedState]);

    useEffect(() => {
        const unblock = history.block((location) => {
            if (displayGetStarted && getStartedStateRef.current.completedSteps < 4) {
                modalFlip(true);
                setTargetLocation(location.pathname);
                return false;
            } else skipGetStarted();
        });
        return () => {
            unblock();
        };
    }, [displayGetStarted, history]);

    const handleConfirm = () => {
        skipGetStarted();
        setDisplayGetStarted(false);
        targetLocation ? history.push(targetLocation) : skip();
        modalFlip(false);
    };

    const skipGetStarted = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, { username: capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_USER_NAME)) });
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, true);
        } catch (error) {}
    };
    const getStepsDescription = (stepNumber) => {
        switch (stepNumber) {
            case 1:
                return 'A station is a distributed unit that stores the produced data';
            case 2:
                return `Each producer/consumer has to have a username and a ${
                    localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' ? 'password' : 'connection-token'
                }`;
            case 3:
                return 'A producer is the source application/service that pushes data or messages to the broker or station';
            case 4:
                return 'A consumer is the application/service that consume data or messages from the broker or station';
            case 5:
                return (
                    <div className="congratulations-section">
                        <label>Congratulations!</label>
                        <label>You've created your first fully-operational station.</label>
                        <label>Continue your journey and connect Memphis with more clients.</label>
                    </div>
                );
        }
    };

    const SideStepList = () => {
        return (
            <div className="sidebar-component">
                {steps.map((value, index) => {
                    return (
                        <SideStep
                            key={index}
                            currentStep={getStartedState?.currentStep}
                            stepNumber={index + 1}
                            stepName={value.stepName}
                            stepsDescription={getStepsDescription(index + 1)}
                            completedSteps={getStartedState?.completedSteps}
                            onSideBarClick={(e) => getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: e })}
                        />
                    );
                })}
            </div>
        );
    };

    const onNext = () => {
        createStationFormRef.current();
    };

    const onBack = () => {
        getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep - 1 });
    };

    const getAvailableReplicas = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_AVAILABLE_REPLICAS);
            getStartedDispatch({ type: 'SET_ACTUAL_PODS', payload: Array.from({ length: data?.available_replicas }, (_, i) => i + 1) });
        } catch (error) {}
    };

    useEffect(() => {
        if (!isCloud()) {
            getAvailableReplicas();
        }
        return () => {
            getStartedDispatch({ type: 'INITIAL_STATE', payload: {} });
        };
    }, []);

    useEffect(() => {
        if (getStartedState?.currentStep !== 1) {
            getStartedDispatch({ type: 'SET_BACK_DISABLE', payload: false });
        } else {
            getStartedDispatch({ type: 'SET_BACK_DISABLE', payload: true });
        }
        return;
    }, [getStartedState?.currentStep]);

    return (
        <GetStartedStoreContext.Provider value={[getStartedState, getStartedDispatch]}>
            <div className="getstarted-container">
                <div className="sidebar-section">
                    <div className="welcome-section">
                        <p className="getstarted-welcome">Welcome, {username}</p>
                        <p className="getstarted-description">{dataSentence}</p>
                    </div>
                    <Divider className="divider">
                        <Button
                            width="120px"
                            height="36px"
                            fontFamily="InterSemiBold"
                            placeholder="Skip for now"
                            radiusType="circle"
                            backgroundColorType="purple"
                            colorType="white"
                            fontSize="14px"
                            boxShadow="gray"
                            onClick={() => {
                                modalFlip(true);
                            }}
                        />
                    </Divider>

                    <div className="getstarted-message-container">
                        <p className="getstarted-message">Let’s get you started</p>
                        <p className="getstarted-message-description">Your streaming journey with Memphis starts here</p>
                    </div>

                    <SideStepList />
                </div>
                <div className="steps-section">
                    {getStartedState?.currentStep === 1 && (
                        <GetStartedItem
                            headerImage={GetStartedIcon}
                            headerTitle="Create Station"
                            headerDescription={getStepsDescription(getStartedState?.currentStep)}
                            onNext={onNext}
                            onBack={onBack}
                        >
                            <CreateStation createStationFormRef={createStationFormRef} />
                        </GetStartedItem>
                    )}
                    {getStartedState?.currentStep === 2 && (
                        <GetStartedItem
                            headerImage={AppUserIcon}
                            headerTitle="Create app user"
                            headerDescription={getStepsDescription(getStartedState?.currentStep)}
                            onNext={onNext}
                            onBack={onBack}
                        >
                            <CreateAppUser createStationFormRef={createStationFormRef} />
                        </GetStartedItem>
                    )}
                    {getStartedState?.currentStep === 3 && (
                        <GetStartedItem
                            headerImage={ProduceDataImg}
                            headerTitle="Produce data"
                            headerDescription={getStepsDescription(getStartedState?.currentStep)}
                            onNext={onNext}
                            onBack={onBack}
                        >
                            <ProduceConsumeData createStationFormRef={createStationFormRef} />
                        </GetStartedItem>
                    )}
                    {getStartedState?.currentStep === 4 && (
                        <GetStartedItem
                            headerImage={ConsumeDataImg}
                            headerTitle="Consume data"
                            headerDescription={getStepsDescription(getStartedState?.currentStep)}
                            onNext={onNext}
                            onBack={onBack}
                        >
                            <ProduceConsumeData consumer={true} createStationFormRef={createStationFormRef} />
                        </GetStartedItem>
                    )}
                    {getStartedState?.currentStep === 5 && (
                        <GetStartedItem
                            headerImage={finishStep}
                            headerTitle="You are ready to stream"
                            headerDescription={getStepsDescription(getStartedState?.currentStep)}
                            onNext={onNext}
                            onBack={onBack}
                            style={finishStyle}
                            finish
                        >
                            <Finish createStationFormRef={createStationFormRef} />
                        </GetStartedItem>
                    )}
                </div>
            </div>
            <SkipGetStrtedModal
                open={open}
                cancel={() => {
                    modalFlip(false);
                }}
                skip={handleConfirm}
            />
        </GetStartedStoreContext.Provider>
    );
};
export const GetStartedStoreContext = createContext({});
export default GetStarted;
