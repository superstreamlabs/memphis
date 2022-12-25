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

import React, { useEffect, useContext, useState } from 'react';

import { Context } from '../../hooks/store';
import SchemaList from './components/schemaList';
import CreateSchema from './components/createSchema';

function SchemaManagment() {
    const [state, dispatch] = useContext(Context);
    const [schemaAction, setSchemaAction] = useState('');

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'schemaverse' });
        const url = window.location.href;
        const schema = url.split('schemaverse/')[1];
        setSchemaAction(schema);
    }, []);

    const createNew = (e) => {
        if (e) setSchemaAction('$new');
        else setSchemaAction('');
    };

    return <div>{schemaAction === '$new' ? <CreateSchema createNew={(e) => createNew(e)} /> : <SchemaList createNew={(e) => createNew(e)} />}</div>;
}

export default SchemaManagment;
