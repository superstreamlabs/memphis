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

import React, { createContext, useEffect, useReducer, useRef } from 'react';

import SideStep from './sideStep';
import CreateAppUser from './createAppUser';
import ConsumeData from './consumeData';
import Reducer from './hooks/reducer';
import ProduceData from './produceData';
import CreateStation from './createStation';

import GetStartedItem from '../../../components/getStartedItem';
import GetStartedIcon from '../../../assets/images/getStartedIcon.svg';
import AppUserIcon from '../../../assets/images/usersIconActive.svg';
import ProduceDataImg from '../../../assets/images/emptyStation.svg';
import ConsumeDataImg from '../../../assets/images/fullStation.svg';
import finishStep from '../../../assets/images/readyToRoll.svg';
import Finish from './finish';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import Button from '../../../components/button';
import { LOCAL_STORAGE_SKIP_GET_STARTED } from '../../../const/localStorageConsts';
import pathDomains from '../../../router';
import { useHistory } from 'react-router-dom';

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
        retention_value: 604800,
        storage_type: 'file',
        replicas: '1',
        days: 7,
        hours: 0,
        minutes: 0,
        seconds: 0,
        retentionSizeValue: '1000',
        retentionMessagesValue: '10'
    },
    nextDisable: false,
    isLoading: false,
    isHiddenButton: false,
    actualPods: null
};

const GetStarted = ({ username, dataSentence }) => {
    const [getStartedState, getStartedDispatch] = useReducer(Reducer, initialState);
    const history = useHistory();
    const createStationFormRef = useRef(null);

    const getStepsDescription = (stepNumber) => {
        switch (stepNumber) {
            case 1:
                return 'A station is a distributed unit that stores the produced data';
            case 2:
                return 'Each producer/consumer has to have a username and a connection-token';
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

    const getOverviewData = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_MAIN_OVERVIEW_DATA);
            let indexOfBrokerComponent = data?.system_components.findIndex((item) => item.component.includes('broker'));
            indexOfBrokerComponent = indexOfBrokerComponent !== -1 ? indexOfBrokerComponent : 1;
            getStartedDispatch({ type: 'SET_ACTUAL_PODS', payload: data?.system_components[indexOfBrokerComponent]?.actual_pods });
        } catch (error) {}
    };

    const skipGetStarted = async (bodyRequest) => {
        try {
            await httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, bodyRequest);
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, true);
            history.push(pathDomains.overview);
        } catch (error) {}
    };

    useEffect(() => {
        getOverviewData();
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
                    <div className="getstarted-message-container">
                        <p className="getstarted-message">Let’s get you started</p>
                        <p className="getstarted-message-description">Your streaming journey with Memphis starts here</p>
                    </div>
                    <SideStepList />
                    <div className="skip-btn">
                        <Button
                            width="120px"
                            height="36px"
                            fontFamily="InterSemiBold"
                            placeholder="Skip for now"
                            radiusType="circle"
                            backgroundColorType="none"
                            border="gray"
                            fontSize="14px"
                            boxShadow="gray"
                            onClick={() => {
                                skipGetStarted({ username });
                            }}
                        />
                    </div>
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
                            <ProduceData createStationFormRef={createStationFormRef} />
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
                            <ConsumeData createStationFormRef={createStationFormRef} />
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
        </GetStartedStoreContext.Provider>
    );
};
export const GetStartedStoreContext = createContext({});
export default GetStarted;
