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

import React, { useEffect, useContext } from 'react';

import { Context } from '../../hooks/store';
import SchemaList from './components/schemaList';
import CreateSchema from './components/createSchema';

function SchemaManagment() {
    const [state, dispatch] = useContext(Context);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'schemas' });
    }, []);

    return <div>{state?.createSchema ? <CreateSchema /> : <SchemaList />}</div>;
}

export default SchemaManagment;
