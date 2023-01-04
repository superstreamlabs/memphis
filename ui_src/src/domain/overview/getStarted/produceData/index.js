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

import React, { useContext, useEffect, useState } from 'react';
import prodWaiting from '../../../../assets/images/waitingForProducersvg.svg';
import ProduceConsumeData, { produceConsumeScreenEnum } from '../produceConsumeData';
import { GetStartedStoreContext } from '..';

const ProduceData = (props) => {
    const { createStationFormRef } = props;
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [displayScreen, setDisplayScreen] = useState();
    const selectLngOption = ['Go', 'Node.js', 'TypeScript', 'Python'];

    const onNext = () => {
        if (displayScreen === produceConsumeScreenEnum['DATA_SNIPPET']) {
            setDisplayScreen(produceConsumeScreenEnum['DATA_WAITING']);
        } else {
            getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
        }
    };

    useEffect(() => {
        createStationFormRef.current = onNext;
    }, [displayScreen]);

    useEffect(() => {
        setDisplayScreen(produceConsumeScreenEnum['DATA_SNIPPET']);
    }, []);

    return (
        <div className="produce-consume-data">
            <ProduceConsumeData
                waitingImage={prodWaiting}
                waitingTitle={
                    <div>
                        <p className="waiting-message">Waiting for produced data</p>
                        <p className="description">
                            Please run the copied code snippet to test your connectivity.
                            <br />
                            Make sure you the broker host address is available to your location
                        </p>
                    </div>
                }
                successfullTitle={'Success! You created your first producer'}
                languages={selectLngOption}
                activeData={'connected_producers'}
                dataName={'producer_app'}
                displayScreen={displayScreen}
                produce
                screen={(e) => setDisplayScreen(e)}
            ></ProduceConsumeData>
        </div>
    );
};

export default ProduceData;
