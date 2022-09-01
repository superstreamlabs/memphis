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

import React, { createContext, useEffect, useReducer, useRef } from 'react';

import CreateStationForm from './createStationForm';
import SideStep from './sideStep';
import CreateAppUser from './createAppUser';
import ConsumeData from './consumeData';
import Reducer from './hooks/reducer';
import ProduceData from './produceData';
import GetStartedItem from '../../../components/getStartedItem';
import GetStartedIcon from '../../../assets/images/getStartedIcon.svg';
import AppUserIcon from '../../../assets/images/usersIconActive.svg';
import ProduceDataImg from '../../../assets/images/produceData.svg';
import ConsumeDataImg from '../../../assets/images/consumeData.svg';
import finishStep from '../../../assets/lotties/finishStep.json';
import Finish from './finish';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';

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
        factory_name: '',
        name: '',
        retention_type: 'message_age_sec',
        retention_value: 604800,
        storage_type: 'file',
        replicas: 1,
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
    desiredPods: null
};

const GetStarted = (props) => {
    const createStationFormRef = useRef(null);
    const [getStartedState, getStartedDispatch] = useReducer(Reducer, initialState);

    const SideStepList = () => {
        return steps.map((value, index) => {
            return (
                <SideStep
                    key={index}
                    currentStep={getStartedState?.currentStep}
                    stepNumber={index + 1}
                    stepName={value.stepName}
                    completedSteps={getStartedState?.completedSteps}
                    onSideBarClick={(e) => getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: e })}
                />
            );
        });
    };

    const onNext = () => {
        createStationFormRef.current();
    };

    const onBack = () => {
        getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep - 1 });
    };

    useEffect(() => {
        getOverviewData();
        return;
    }, []);

    useEffect(() => {
        if (getStartedState?.currentStep !== 1) {
            getStartedDispatch({ type: 'SET_BACK_DISABLE', payload: false });
        } else {
            getStartedDispatch({ type: 'SET_BACK_DISABLE', payload: true });
        }
        return;
    }, [getStartedState?.currentStep]);

    const getOverviewData = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_MAIN_OVERVIEW_DATA);
            getStartedDispatch({ type: 'SET_DESIRED_PODS', payload: data?.system_components[1]?.desired_pods });
        } catch (error) {}
    };

    return (
        <GetStartedStoreContext.Provider value={[getStartedState, getStartedDispatch]}>
            <div className="getstarted-container">
                <h1 className="getstarted-header">Let's get you started</h1>
                <p className="getstarted-header-description">Setup your account details to get more from the platform</p>
                <div className="sub-getstarted-container">
                    <div className="side-step">
                        <SideStepList />
                    </div>
                    <div className="getstarted-item">
                        {getStartedState?.currentStep === 1 && (
                            <GetStartedItem
                                headerImage={GetStartedIcon}
                                headerTitle="Create Station"
                                headerDescription="Station is the object that stores data"
                                onNext={onNext}
                                onBack={onBack}
                            >
                                <CreateStationForm createStationFormRef={createStationFormRef} />
                            </GetStartedItem>
                        )}
                        {getStartedState?.currentStep === 2 && (
                            <GetStartedItem
                                headerImage={AppUserIcon}
                                headerTitle="Create application user"
                                headerDescription="User of type application is for connecting apps"
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
                                headerDescription="Choose your preferred SDK, copy and paste the code to your IDE, and run your app to produce data into a memphis station"
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
                                headerDescription="Choose your preferred SDK, copy and paste the code to your IDE, and run your app to consume data from memphis station"
                                onNext={onNext}
                                onBack={onBack}
                            >
                                <ConsumeData createStationFormRef={createStationFormRef} />
                            </GetStartedItem>
                        )}
                        {getStartedState?.currentStep === 5 && (
                            <GetStartedItem
                                headerImage={finishStep}
                                headerTitle="You are ready to roll"
                                headerDescription="Congratulations - You’ve created your first broker app"
                                onNext={onNext}
                                onBack={onBack}
                                style={finishStyle}
                            >
                                <Finish createStationFormRef={createStationFormRef} />
                            </GetStartedItem>
                        )}
                    </div>
                </div>
            </div>
        </GetStartedStoreContext.Provider>
    );
};
export const GetStartedStoreContext = createContext({});
export default GetStarted;
