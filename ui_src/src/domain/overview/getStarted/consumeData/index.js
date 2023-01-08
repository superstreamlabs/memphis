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

import React, { useContext, useEffect, useState } from 'react';
import consWaiting from '../../../../assets/images/waitingForConsumer.svg';
import ProduceConsumeData, { produceConsumeScreenEnum } from '../produceConsumeData';
import { GetStartedStoreContext } from '..';

const ConsumeData = (props) => {
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
                waitingImage={consWaiting}
                waitingTitle={
                    <div>
                        <p className="waiting-message">Waiting to consume messages from the station</p>
                        <p className="description">
                            Please run the copied code snippet to test your connectivity.
                            <br />
                            Make sure you the broker host address is available to your location
                        </p>
                    </div>
                }
                successfullTitle={'Success! You created your first consumer'}
                languages={selectLngOption}
                activeData={'connected_cgs'}
                dataName={'consumer_app'}
                displayScreen={displayScreen}
                consume
                screen={(e) => setDisplayScreen(e)}
            ></ProduceConsumeData>
        </div>
    );
};

export default ConsumeData;
