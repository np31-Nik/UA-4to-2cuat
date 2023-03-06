# IO.Swagger.Api.DefaultApi

All URIs are relative to *https://virtserver.swaggerhub.com/SIG4/HolaMundo/1.0.0*

Method | HTTP request | Description
------------- | ------------- | -------------
[**HolamundoGet**](DefaultApi.md#holamundoget) | **GET** /holamundo | 
[**HolamundoPut**](DefaultApi.md#holamundoput) | **PUT** /holamundo | 

<a name="holamundoget"></a>
# **HolamundoGet**
> InlineResponse200 HolamundoGet (string nombreHolaMundo)



GET Hola Mundo

### Example
```csharp
using System;
using System.Diagnostics;
using IO.Swagger.Api;
using IO.Swagger.Client;
using IO.Swagger.Model;

namespace Example
{
    public class HolamundoGetExample
    {
        public void main()
        {
            var apiInstance = new DefaultApi();
            var nombreHolaMundo = nombreHolaMundo_example;  // string | nombre Hola Mundo

            try
            {
                InlineResponse200 result = apiInstance.HolamundoGet(nombreHolaMundo);
                Debug.WriteLine(result);
            }
            catch (Exception e)
            {
                Debug.Print("Exception when calling DefaultApi.HolamundoGet: " + e.Message );
            }
        }
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nombreHolaMundo** | **string**| nombre Hola Mundo | 

### Return type

[**InlineResponse200**](InlineResponse200.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
<a name="holamundoput"></a>
# **HolamundoPut**
> void HolamundoPut (string nombreHolaMundo)



POST Hola Mundo

### Example
```csharp
using System;
using System.Diagnostics;
using IO.Swagger.Api;
using IO.Swagger.Client;
using IO.Swagger.Model;

namespace Example
{
    public class HolamundoPutExample
    {
        public void main()
        {
            var apiInstance = new DefaultApi();
            var nombreHolaMundo = nombreHolaMundo_example;  // string | nombre Hola Mundo

            try
            {
                apiInstance.HolamundoPut(nombreHolaMundo);
            }
            catch (Exception e)
            {
                Debug.Print("Exception when calling DefaultApi.HolamundoPut: " + e.Message );
            }
        }
    }
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **nombreHolaMundo** | **string**| nombre Hola Mundo | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)
