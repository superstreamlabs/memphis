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

import React, { useContext, useEffect } from 'react';
import { GetStartedStoreContext } from '..';
import CreateStationForm from '../../../../components/createStationForm';

import './style.scss';

const CreateStation = ({ createStationFormRef }) => {
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);

    useEffect(() => {
        getStartedState?.completedSteps === 0
            ? getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: true })
            : getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
    }, []);

    const createStationDone = (data) => {
        if (data) {
            getStartedDispatch({ type: 'SET_STATION', payload: data.name });
            getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
        } else {
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
        }
    };

    const updateFormState = (field, value) => {
        if (field === 'name') {
            value.length > 0 ? getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false }) : getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: true });
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
