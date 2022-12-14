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
