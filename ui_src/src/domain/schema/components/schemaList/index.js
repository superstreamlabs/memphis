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

import React, { useEffect, useContext, useState } from 'react';
import { SearchOutlined } from '@ant-design/icons';

import emptyList from '../../../../assets/images/emptyList.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import SearchInput from '../../../../components/searchInput';
import { httpRequest } from '../../../../services/http';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import { Context } from '../../../../hooks/store';
import SchemaBox from '../schemaBox';

function SchemaList({ createNew }) {
    const [state, dispatch] = useContext(Context);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [schemaList, setSchemaList] = useState([
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock1',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                },
                {
                    id: 3,
                    active: false,
                    version_number: 3,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 4,
                    active: false,
                    version_number: 4,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock2',
            used: true,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1_gfdgfdgfdgfgd', 'station_2_jghghhjhj', 'station_3']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock3',
            used: true,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: true,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: false,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock1',
            used: true,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        },
        {
            id: 'jhjhjkhkjhkj',
            name: 'schema_mock',
            used: true,
            type: 'avro',
            created_by: 'avraham',
            tags: [
                { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
                { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
                { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
                { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
            ],
            creation_date: '2022-09-21T14:28:02.539503+03:00',
            versions: [
                {
                    id: 1,
                    active: false,
                    version_number: 1,
                    name: 'schema_mock2',
                    created_by_user: 'avraham',
                    schema: `{
    "type": "record",
    "namespace": "com.example",
    "name": "test-schema",
    "fields": [
       { "name": "username", "type": "string", "default": "-2" },
       { "name": "age", "type": "int", "default": "none" },
       { "name": "phone", "type": "int", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
    ]
}`
                },
                {
                    id: 2,
                    active: true,
                    version_number: 2,
                    name: 'schema_mock1',
                    created_by_user: 'avraham',
                    schema: `{
                                "type": "record",
                                "namespace": "com.example",
                                "name": "test-schema",
                                "fields": [
                                { "name": "username", "type": "string", "default": "NONE" },
                                { "name": "age", "type": "int", "default": "-1" },
                                { "name": "phone", "type": "string", "default": "NONE" },
                                { "name": "country", "type": "string", "default": "NONE" }
                                ]
                            }`
                }
            ],
            stations: ['station_1', 'station_2', 'station_3', 'station_4']
        }
    ]);
    const [isLoading, setisLoading] = useState(false);

    const getSchemas = async () => {
        // setisLoading(true);
        // try {
        //     debugger;
        //     const data = await httpRequest('GET', ApiEndpoints.GEL_ALL_FACTORIES);
        //     setSchemaList(data);
        //     setisLoading(false);
        // } catch (error) {
        //     setisLoading(false);
        // }
    };

    useEffect(() => {
        getSchemas();
    }, []);

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(schemaList.map((li) => li.id));
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
        if (isCheck.length === 1 && !checked) {
            setIsCheckAll(false);
        }
    };
    return (
        <div className="schema-container">
            <h1 className="main-header-h1">Schema</h1>
            <div className="action-section">
                <SearchInput
                    placeholder="Search schema"
                    colorType="navy"
                    backgroundColorType="gray-dark"
                    width="288px"
                    height="34px"
                    borderRadiusType="circle"
                    borderColorType="none"
                    boxShadowsType="none"
                    iconComponent={<SearchOutlined />}
                    // onChange={handleSearch}
                    // value={searchInput}
                />
                <Button
                    width="111px"
                    height="34px"
                    placeholder={'Filters'}
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="12px"
                    fontWeight="600"
                    aria-haspopup="true"
                    // onClick={() => addUserModalFlip(true)}
                />
                {/* <Button
                    width="81px"
                    height="34px"
                    placeholder={'Sort'}
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="12px"
                    fontWeight="600"
                    aria-haspopup="true"
                    // onClick={() => addUserModalFlip(true)}
                /> */}
                <Button
                    width="160px"
                    height="34px"
                    placeholder={'Create from blank'}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="600"
                    aria-haspopup="true"
                    onClick={() => createNew()}
                />
                {/* <Button
                    width="145px"
                    height="34px"
                    placeholder={'Import schema'}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="600"
                    aria-haspopup="true"
                    // onClick={() => createNew()}
                /> */}
            </div>
            <div className="schema-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {schemaList.map((schema, index) => {
                    return <SchemaBox key={index} schema={schema} isCheck={isCheck.includes(schema.id)} handleCheckedClick={handleCheckedClick} />;
                })}
                {!isLoading && schemaList.length === 0 && (
                    <div className="no-schema-to-display">
                        <img src={emptyList} width="100" height="100" alt="emptyList" />
                        <p>There are no schema yet</p>
                        <p className="sub-title">Get started by creating your first schema</p>
                        <Button
                            className="modal-btn"
                            width="240px"
                            height="50px"
                            placeholder="Create your first schema"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            // onClick={() => modalFlip(true)}
                        />
                    </div>
                )}
            </div>
        </div>
    );
}

export default SchemaList;
