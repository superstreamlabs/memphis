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

import React, { useContext, useEffect, useState } from 'react';
import consWaiting from '../../../../assets/lotties/consWaiting.json';
import ProduceConsumeData, { produceConsumeScreenEnum } from '../produceConsumeData';
import { GetStartedStoreContext } from '..';

const ConsumeData = (props) => {
    const { createStationFormRef } = props;
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [displayScreen, setDisplayScreen] = useState();
    const selectLngOption = ['Go', 'Node.js', 'Typescript', 'Python'];

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
        <ProduceConsumeData
            waitingImage={consWaiting}
            waitingTitle={'Waiting to consume messages from the station'}
            successfullTitle={'Success! You created your first consumer'}
            languages={selectLngOption}
            activeData={'connected_cgs'}
            dataName={'consumer_app'}
            displayScreen={displayScreen}
            consume
            screen={(e) => setDisplayScreen(e)}
        ></ProduceConsumeData>
    );
};

export default ConsumeData;
