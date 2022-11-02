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

import React, { useEffect, useContext, useState, useReducer, createContext } from 'react';

import { Context } from '../../hooks/store';
import SchemaList from './components/schemaList';
import CreateSchema from './components/createSchema';

function SchemaManagment() {
    const [state, dispatch] = useContext(Context);
    const [creatNew, setCreatNew] = useState(false);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'schemas' });
    }, []);

    return <div>{creatNew ? <CreateSchema goBack={() => setCreatNew(false)} /> : <SchemaList createNew={() => setCreatNew(true)} />}</div>;
}

export default SchemaManagment;
