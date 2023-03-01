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

import React, { useContext, useEffect } from 'react';
import { GetStartedStoreContext } from '..';
import CreateStationForm from '../../../../components/createStationForm';
import { generateName } from '../../../../services/valueConvertor';

import './style.scss';

const CreateStation = ({ createStationFormRef }) => {
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);

    useEffect(() => {
        getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: getStartedState?.completedSteps === 0 || false });
    }, []);

    const createStationDone = (data) => {
        if (data) {
            getStartedDispatch({ type: 'SET_STATION', payload: data.name });
            getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: true });
        } else {
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: getStartedState?.username ? false : true });
        }
    };

    const updateFormState = (field, value) => {
        if (field === 'name') {
            value = generateName(value);
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: value.length === 0 || false });
        }
        getStartedDispatch({ type: 'SET_FORM_FIELDS_CREATE_STATION', payload: { field: field, value: value } });
    };

    const setLoading = (payload) => {
        getStartedDispatch({ type: 'IS_LOADING', payload: payload });
    };

    return (
        <CreateStationForm
            getStarted
            getStartedStateRef={getStartedState}
            createStationFormRef={createStationFormRef}
            updateFormState={(field, value) => updateFormState(field, value)}
            setLoading={setLoading}
            finishUpdate={(e) => createStationDone(e)}
        />
    );
};
export default CreateStation;
